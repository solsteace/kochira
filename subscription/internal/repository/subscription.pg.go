package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/solsteace/kochira/subscription/internal/domain"
)

type pgSubscriptionRow struct {
	Id        uint64    `db:"id"`
	UserId    uint64    `db:"user_id"`
	ExpiredAt time.Time `db:"expired_at"`
}

func (row pgSubscriptionRow) ToDomain() (domain.Subscription, error) {
	return domain.NewSubscription(
		&row.Id,
		row.UserId,
		row.ExpiredAt)
}

func newPgSubscriptionRow(s domain.Subscription) pgSubscriptionRow {
	return pgSubscriptionRow{
		Id:        s.Id(),
		UserId:    s.UserId(),
		ExpiredAt: s.ExpiredAt()}
}

type pgSubscription struct {
	db *sqlx.DB
}

func NewPgSubscription(db *sqlx.DB) pgSubscription {
	return pgSubscription{db}
}

func (repo pgSubscription) GetByOwner(id uint64) (domain.Subscription, error) {
	row := new(pgSubscriptionRow)
	query := `
		SELECT *
		FROM subscriptions
		WHERE user_id = $1`
	args := []any{id}
	if err := repo.db.Get(row, query, args...); err != nil {
		return domain.Subscription{}, err
	}

	return row.ToDomain()
}

func (repo pgSubscription) Create(s domain.Subscription) error {
	return nil
}

func (repo pgSubscription) Update(s domain.Subscription) error {
	return nil
}

func (repo pgSubscription) Delete(id uint64) error {
	return nil
}
