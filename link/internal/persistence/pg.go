package persistence

import "github.com/jmoiron/sqlx"

type pg struct {
	db *sqlx.DB
}

func NewPgLink(db *sqlx.DB) pg {
	return pg{db}
}
