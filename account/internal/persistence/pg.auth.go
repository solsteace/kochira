package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/account/internal/domain/auth"
)

type pgAuth struct {
	db *sqlx.DB
}

func NewPgAuth(db *sqlx.DB) pgAuth {
	return pgAuth{db: db}
}

type pgAuthUser struct {
	Id       uint   `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
}

func (row pgAuthUser) ToUser() (auth.User, error) {
	return auth.NewUser(
		&row.Id,
		row.Username,
		row.Password)
}

func newPgAuthUser(a auth.User) pgAuthUser {
	return pgAuthUser{
		a.Id,
		a.Username,
		a.Password}
}

func (repo pgAuth) GetByUsername(username string) (auth.User, error) {
	row := new(pgAuthUser)
	query := `
		SELECT id, username, password
		FROM users WHERE username = $1`
	args := []any{username}
	if err := repo.db.Get(row, query, args...); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return auth.User{}, fmt.Errorf(
				"persistence<pgAuth.GetByUsername>: %w",
				oops.NotFound{
					Err: err,
					Msg: fmt.Sprintf("user(username:%s) not found", username)})
		default:
			return auth.User{}, fmt.Errorf("persistence<pgAuth.GetByUsername>: %w", err)
		}
	}

	user, err := row.ToUser()
	if err != nil {
		return auth.User{}, fmt.Errorf("persistence<pgAuth.GetById>: %w", err)
	}
	return user, nil
}
