package persistence

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/messaging"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/value"
)

type pgSubscription struct {
	Id        uint64    `db:"id"`
	UserId    uint64    `db:"user_id"`
	ExpiredAt time.Time `db:"expired_at"`
}

func (row pgSubscription) ToDomain() (subscription.Subscription, error) {
	return subscription.NewSubscription(
		&row.Id,
		row.UserId,
		row.ExpiredAt)
}

func newPgSubscriptionRow(s subscription.Subscription) pgSubscription {
	return pgSubscription{
		Id:        s.Id(),
		UserId:    s.UserId(),
		ExpiredAt: s.ExpiredAt()}
}

func (repo pg) FilterExisting(id []uint64) ([]uint64, error) {
	// Warning: SQL standard, but not universal in the way of adoption
	idRows := []string{}
	for _, i := range id {
		idRows = append(idRows, fmt.Sprintf("(%d)", i))
	}
	idToCheck := fmt.Sprintf("VALUES %s", strings.Join(idRows, ","))

	query := fmt.Sprintf(`
		WITH id_to_check("id") AS (%s)
		SELECT id FROM id_to_check
		WHERE id NOT IN (SELECT id FROM subscriptions)`, idToCheck)
	filteredId := new([]uint64)
	if err := repo.db.Select(filteredId, query); err != nil {
		return []uint64{}, fmt.Errorf(
			"persistence<pg.CheckManyByOwner>: %w", err)
	}
	return *filteredId, nil
}

func (repo pg) GetByOwner(id uint64) (subscription.Subscription, error) {
	row := new(pgSubscription)
	query := `
		SELECT
			id, 
			user_id,
			expired_at
		FROM subscriptions
		WHERE user_id = $1`
	args := []any{id}
	if err := repo.db.Get(row, query, args...); err != nil {
		return subscription.Subscription{}, fmt.Errorf(
			"persistence<pg.GetByOwner>: %w", err)
	}

	return row.ToDomain()
}

func (repo pg) Create(subscriptions []subscription.Subscription) error {
	rows := []pgSubscription{}
	for _, s := range subscriptions {
		rows = append(rows, newPgSubscriptionRow(s))
	}

	query := `
		INSERT INTO subscriptions(user_id, expired_at)
		VALUES (:user_id, :expired_at)`
	if _, err := repo.db.NamedExec(query, rows); err != nil {
		return fmt.Errorf("persistence<pg.Create>: %w", err)
	}
	return nil
}

// Events ======================

type pgSubscriptionChecked struct {
	Id             uint64        `db:"id"`
	ContextId      uint64        `db:"context_id"`
	Usecase        string        `db:"usecase"`
	Lifetime       time.Duration `db:"lifetime"`
	Limit          uint          `db:"limit"`
	AllowShortEdit bool          `db:"allow_short_edit"`
}

func (row pgSubscriptionChecked) toMsg() messaging.SubscriptionChecked {
	return messaging.NewSubscriptionChecked(
		row.Id,
		row.ContextId,
		row.Usecase,
		row.Lifetime,
		row.Limit,
		row.AllowShortEdit)
}

func (repo pg) CreateSubscriptionChecked(
	contextId uint64,
	usecase string,
	perk value.Perk,
) error {
	query := `
		INSERT INTO subscription_checked_outbox(
			context_id,
			usecase,
			lifetime, 
			"limit", 
			allow_short_edit)
		VALUES ($1, $2, $3, $4, $5)`
	args := []any{
		contextId,
		usecase,
		perk.Lifetime(),
		perk.Limit(),
		perk.AllowShortEdit()}
	if _, err := repo.db.Exec(query, args...); err != nil {
		return fmt.Errorf("persistence<pg.CreateSubscriptionChecked>: %w", err)
	}
	return nil
}

func (repo pg) GetSubscriptionChecked(limit uint) ([]messaging.SubscriptionChecked, error) {
	query := `
		SELECT 
			id, 
			context_id,
			usecase,
			lifetime, 
			"limit", 
			allow_short_edit 
		FROM subscription_checked_outbox
		WHERE is_done = false
		LIMIT $1`
	args := []any{limit}
	rows := new([]pgSubscriptionChecked)
	if err := repo.db.Select(rows, query, args...); err != nil {
		return []messaging.SubscriptionChecked{}, fmt.Errorf(
			"persistence<pg.GetSubscriptionChecked>: %w", err)
	}

	msg := []messaging.SubscriptionChecked{}
	for _, r := range *rows {
		msg = append(msg, r.toMsg())
	}
	return msg, nil
}

func (repo pg) ResolveSubscriptionChecked(id []uint64) error {
	query, args, err := sqlx.In(`
		UPDATE subscription_checked_outbox
		SET is_done = true
		WHERE id IN (?)`, id)
	if err != nil {
		return fmt.Errorf("persistence<pg.ResolveSubscriptionChecked>: %w", err)
	}

	if _, err := repo.db.Exec(repo.db.Rebind(query), args...); err != nil {
		return fmt.Errorf("persistence<pg.ResolveSubscriptionChecked>: %w", err)
	}
	return nil
}

type pgSubscriptionExpired struct {
	Id     uint64 `db:"id"`
	UserId uint64 `db:"user_id"`
}

func (row pgSubscriptionExpired) ToMsg() messaging.SubscriptionExpired {
	return messaging.NewSubscriptionExpired(row.Id, row.UserId)
}

func (repo pg) WatchExpiringSubscription(limit uint) error {
	expiredSubscription := new([]pgSubscriptionExpired)
	query := `
		SELECT 
			id,
			user_id
		FROM subscriptions
		WHERE 
			expired_at < CURRENT_TIMESTAMP
			AND checked_at IS NULL
		LIMIT $1`
	args := []any{limit}
	if err := repo.db.Select(expiredSubscription, query, args...); err != nil {
		return fmt.Errorf("persistence<pg.WatchExpiringSubscription>: %w", err)
	} else if len(*expiredSubscription) == 0 {
		return nil
	}

	ctx, _ := context.WithCancel(context.Background())
	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("persistence<pg.WatchExpiringSubscription>: %w", err)
	}
	defer tx.Rollback()

	query = `
		INSERT INTO subscription_expired_outbox(user_id)
		VALUES (:user_id)`
	if _, err := tx.NamedExec(query, *expiredSubscription); err != nil {
		return fmt.Errorf("persistence<pg.WatchExpiringSubscription>: %w", err)
	}

	userId := []uint64{}
	for _, s := range *expiredSubscription {
		userId = append(userId, s.UserId)
	}
	query, args, err = sqlx.In(`
		UPDATE subscriptions
		SET checked_at = CURRENT_TIMESTAMP
		WHERE user_id IN (?)`, userId)
	if err != nil {
		return fmt.Errorf("persistence<pg.WatchExpiringSubscription>: %w", err)
	}
	if _, err := tx.Exec(tx.Rebind(query), args...); err != nil {
		return fmt.Errorf("persistence<pg.WatchExpiringSubscription>: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("persistence<pg.WatchExpiringSubscription>: %w", err)
	}
	return nil
}

func (repo pg) GetSubscriptionExpired(limit uint) ([]messaging.SubscriptionExpired, error) {
	query := `
		SELECT id, user_id
		FROM subscription_expired_outbox
		WHERE NOT is_done
		LIMIT $1`
	rows := new([]pgSubscriptionExpired)
	args := []any{limit}
	if err := repo.db.Select(rows, query, args...); err != nil {
		return []messaging.SubscriptionExpired{}, fmt.Errorf(
			"persistence<pg.GetSubscriptionExpired>: %w", err)
	}

	msg := []messaging.SubscriptionExpired{}
	for _, r := range *rows {
		msg = append(msg, r.ToMsg())
	}
	return msg, nil
}

func (repo pg) ResolveSubscriptionExpired(id []uint64) error {
	query, args, err := sqlx.In(`
		UPDATE subscription_expired_outbox
		SET is_done = true
		WHERE id IN (?)`, id)
	if err != nil {
		return fmt.Errorf("persistence<pg.ResolveSubscriptionExpired>: %w", err)
	}
	if _, err := repo.db.Exec(repo.db.Rebind(query), args...); err != nil {
		return fmt.Errorf("persistence<pg.ResolveSubscriptionExpired>: %w", err)
	}
	return nil
}
