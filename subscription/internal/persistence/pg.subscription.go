package persistence

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription"
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
			"persistence<pgSubscription.CheckManyByOwner>: %w", err)
	}

	foundId := new([]uint64)
	if err := repo.db.Select(foundId, repo.db.Rebind(query), args...); err != nil {
		return []uint64{}, fmt.Errorf(
			"persistence<pgSubscription.CheckManyByOwner>: %w", err)
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
			"persistence<pgSubscription.GetByOwner>: %w", err)
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
		return fmt.Errorf("persistence<pgSubscription.Create>: %w", err)
	}
	return nil
}

func (repo pg) Update(s subscription.Subscription) error {
	return nil
}

func (repo pg) Delete(id uint64) error {
	return nil
}
