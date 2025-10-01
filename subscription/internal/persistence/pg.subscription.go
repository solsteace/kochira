package persistence

import (
	"fmt"
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

func (repo pg) CheckManyByOwner(id []uint64) ([]uint64, error) {
	query, args, err := sqlx.In(`
		SELECT user_id 
		FROM subscriptions
		WHERE user_id IN (?)`, id)
	if err != nil {
		return []uint64{}, fmt.Errorf(
			"persistence<pg.CheckManyByOwner>: %w", err)
	}

	foundId := new([]uint64)
	if err := repo.db.Select(foundId, repo.db.Rebind(query), args...); err != nil {
		return []uint64{}, fmt.Errorf(
			"persistence<pg.CheckManyByOwner>: %w", err)
	}

	return *foundId, nil
}

func (repo pg) GetByOwner(id uint64) (subscription.Subscription, error) {
	row := new(pgSubscription)
	query := `
		SELECT *
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

func (repo pg) Update(s subscription.Subscription) error {
	return nil
}

func (repo pg) Delete(id uint64) error {
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
