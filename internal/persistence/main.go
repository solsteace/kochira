package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DbConnection[connT any] struct {
	Conn connT
}

func NewPgConnection(url string) (DbConnection[*pgxpool.Pool], error) {
	conn, err := pgxpool.New(context.Background(), url)
	if err != nil {
		msg := fmt.Sprintf("Failed connecting to %s", url)
		return DbConnection[*pgxpool.Pool]{}, errors.New(msg)
	}
	return DbConnection[*pgxpool.Pool]{Conn: conn}, nil
}
