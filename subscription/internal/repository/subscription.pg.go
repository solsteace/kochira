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

func (repo pgSubscription) CheckManyByOwner(id []uint64) ([]uint64, error) {
	query, args, err := sqlx.In(`
		SELECT user_id 
		FROM subscriptions
		WHERE user_id IN (?)`, id)
	if err != nil {
		return []uint64{}, err
	}

	foundId := new([]uint64)
	if err := repo.db.Select(foundId, repo.db.Rebind(query), args...); err != nil {
		return []uint64{}, err
	}

	return *foundId, nil
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

func (repo pgSubscription) Create(subscriptions []domain.Subscription) error {
	rows := []pgSubscriptionRow{}
	for _, s := range subscriptions {
		rows = append(rows, newPgSubscriptionRow(s))
	}

	query := `
		INSERT INTO subscriptions(user_id, expired_at)
		VALUES (:user_id, :expired_at)`
	if _, err := repo.db.NamedExec(query, rows); err != nil {
		return err
	}
	return nil
}

func (repo pgSubscription) Update(s domain.Subscription) error {
	return nil
}

func (repo pgSubscription) Delete(id uint64) error {
	return nil
}
