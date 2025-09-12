package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/account/internal/domain"
	"github.com/solsteace/kochira/account/internal/domain/outbox"
)

type pgUser struct {
	db *sqlx.DB
}

func NewPgUser(db *sqlx.DB) pgUser {
	return pgUser{db}
}

type pgUserRow struct {
	Id       uint   `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Email    string `db:"email"`
}

func (row pgUserRow) ToUser() (domain.User, error) {
	return domain.NewUser(
		&row.Id,
		row.Username,
		row.Password,
		row.Email)
}

func newPgUserRow(a domain.User) pgUserRow {
	return pgUserRow{
		a.Id,
		a.Username,
		a.Password,
		a.Email}
}

type pgRegistrationOutboxRow struct {
	Id     uint64 `db:"id"`
	UserId uint64 `db:"user_id"`
	IsDone bool   `db:"is_done"`
}

func (row pgRegistrationOutboxRow) toOutbox() outbox.Register {
	return outbox.NewRegister(row.Id, row.UserId, row.IsDone)
}

func newPgRegistrationOutboxRow(id uint64, userId uint64, isDone bool) pgRegistrationOutboxRow {
	return pgRegistrationOutboxRow{id, userId, isDone}
}

// ==============================
// Repo implementations
// ==============================

func (repo pgUser) GetById(id uint) (domain.User, error) {
	row := new(pgUserRow)
	query := `SELECT * FROM users WHERE id = $1 OR 1 = 1`
	args := []any{id}
	if err := repo.db.Get(row, query, args...); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return domain.User{}, oops.NotFound{
				Err: err,
				Msg: fmt.Sprintf("user(id:%d) not found", id)}
		default:
			return domain.User{}, err
		}
	}
	return row.ToUser()
}

func (repo pgUser) GetByUsername(username string) (domain.User, error) {
	row := new(pgUserRow)
	query := `SELECT * FROM users WHERE username = $1`
	args := []any{username}
	if err := repo.db.Get(row, query, args...); err != nil {
		return domain.User{}, err
	}
	return row.ToUser()
}

func (repo pgUser) Create(a domain.User) error {
	ctx := context.Background() // Change later
	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	row := newPgUserRow(a)
	stmt, err := tx.PrepareNamed(`
		INSERT INTO users(username, password, email)
		VALUES (:username, :password, :email)
		RETURNING id`)
	if err != nil {
		return err
	}
	var userId uint64
	if err := stmt.Get(&userId, row); err != nil {
		return err
	}

	query := `
		INSERT INTO register_outbox(user_id, is_done)
		VALUES ($1, $2)`
	args := []any{userId, false}
	if _, err := tx.Exec(query, args...); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo pgUser) Update(a domain.User) error {
	row := newPgUserRow(a)
	query := `
		UPDATE users 
		SET
			username = :username,
			password = :password,
			email = :email
		WHERE id = :id`
	if _, err := repo.db.NamedExec(query, row); err != nil {
		return err
	}
	return nil
}

func (repo pgUser) GetRegisterOutbox(count uint) ([]outbox.Register, error) {
	rows := new([]pgRegistrationOutboxRow)
	query := `
		SELECT *
		FROM register_outbox
		WHERE is_done = false
		LIMIT $1 `
	args := []any{count}
	if err := repo.db.Select(rows, query, args...); err != nil {
		return []outbox.Register{}, err
	}

	outbox := []outbox.Register{}
	for _, r := range *rows {
		outbox = append(outbox, r.toOutbox())
	}
	return outbox, nil
}

func (repo pgUser) ResolveRegisterOutbox(id []uint64) error {
	query, args, err := sqlx.In(`
		UPDATE register_outbox
		SET is_done = true
		WHERE user_id IN (?)`, id)
	if err != nil {
		return err
	}

	if _, err = repo.db.Exec(repo.db.Rebind(query), args...); err != nil {
		return err
	}
	return nil
}
