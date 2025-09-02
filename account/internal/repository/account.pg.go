package repository

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/account/internal/domain"
)

type pgUserRow struct {
	Id       uint   `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Email    string `db:"email"`
}

func (row pgUserRow) ToModel() domain.User {
	return domain.User{
		Id:       row.Id,
		Username: row.Username,
		Password: row.Password,
		Email:    row.Email}
}

func newPgAccountRow(a domain.User) pgUserRow {
	return pgUserRow{
		a.Id,
		a.Username,
		a.Password,
		a.Email}
}

type pgUser struct {
	db *sqlx.DB
}

func NewPgAccount(db *sqlx.DB) pgUser {
	return pgUser{db}
}

func (repo pgUser) GetById(id uint) (domain.User, error) {
	row := new(pgUserRow)

	query := "SELECT * FROM users WHERE id = $1"
	args := []any{id}
	if err := repo.db.Select(&row, query, args); err != nil {
		return domain.User{}, nil
	}

	return domain.User{}, nil
}

func (repo pgUser) GetByUsername(username string) (domain.User, error) {
	rows := new([]pgUserRow)

	query := "SELECT * FROM users WHERE username = $1 LIMIT 1"
	if err := repo.db.Select(rows, query, username); err != nil {
		return domain.User{}, err
	}

	if len(*rows) == 0 {
		return domain.User{}, oops.NotFound{
			Err: errors.New(fmt.Sprintf("user with username(%s) not found", username)),
			Msg: "user not found"}
	}
	return (*rows)[0].ToModel(), nil
}

func (repo pgUser) Create(a domain.User) error {
	row := newPgAccountRow(a)

	query := `
		INSERT INTO users(
			username,
			password,
			email)
		VaLUES (
			:username,
			:password,
			:email) `
	if _, err := repo.db.NamedExec(query, row); err != nil {
		return err
	}

	return nil
}

func (repo pgUser) Update(a domain.User) error {
	row := newPgAccountRow(a)

	query := `
		UPDATE users 
		SET
			username = :username,
			password = :password,
			email = :email
		WHERE
			id = :id `
	if _, err := repo.db.NamedExec(query, row); err != nil {
		return err
	}

	return nil
}

func (repo pgUser) DeleteById(id uint) error {
	return nil
}
