package persistence

import "github.com/jmoiron/sqlx"

type pg struct {
	db *sqlx.DB
}

func NewPgSubscription(db *sqlx.DB) pg {
	return pg{db}
}
