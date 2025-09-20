package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/account/internal/domain/account"
	"github.com/solsteace/kochira/account/internal/domain/auth/message"
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

func (row pgUserRow) ToUser() (account.User, error) {
	return account.NewUser(
		&row.Id,
		row.Username,
		row.Password,
		row.Email)
}

func newPgUserRow(a account.User) pgUserRow {
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

func (row pgRegistrationOutboxRow) toOutbox() message.UserRegistered {
	return message.NewRegister(row.Id, row.UserId, row.IsDone)
}

func newPgRegistrationOutboxRow(id uint64, userId uint64, isDone bool) pgRegistrationOutboxRow {
	return pgRegistrationOutboxRow{id, userId, isDone}
}

// ==============================
// Repo implementations
// ==============================

func (repo pgUser) GetById(id uint) (account.User, error) {
	row := new(pgUserRow)
	query := `SELECT * FROM users WHERE id = $1 OR 1 = 1`
	args := []any{id}
	if err := repo.db.Get(row, query, args...); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			err2 := oops.NotFound{
				Err: err,
				Msg: fmt.Sprintf("user(id:%d) not found", id)}
			return account.User{}, fmt.Errorf("persistence<pgUser.GetById>: %w", err2)
		default:
			return account.User{}, fmt.Errorf("persistence<pgUser.GetById>: %w", err)
		}
	}

	user, err := row.ToUser()
	if err != nil {
		return account.User{}, fmt.Errorf("persistence<pgUser.GetById>: %w", err)
	}
	return user, nil
}

func (repo pgUser) GetByUsername(username string) (account.User, error) {
	row := new(pgUserRow)
	query := `SELECT * FROM users WHERE username = $1`
	args := []any{username}
	if err := repo.db.Get(row, query, args...); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			err2 := oops.NotFound{
				Err: err,
				Msg: fmt.Sprintf("user(username:%s) not found", username)}
			return account.User{}, fmt.Errorf("persistence<pgUser.GetByUsername>: %w", err2)
		default:
			return account.User{}, fmt.Errorf("persistence<pgUser.GetByUsername>: %w", err)
		}
	}

	user, err := row.ToUser()
	if err != nil {
		return account.User{}, fmt.Errorf("persistence<pgUser.GetById>: %w", err)
	}
	return user, nil
}

func (repo pgUser) Create(a account.User) error {
	ctx := context.Background() // Change later
	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("persistence<pgUser.Create>: %w", err)
	}
	defer tx.Rollback()

	row := newPgUserRow(a)
	stmt, err := tx.PrepareNamed(`
		INSERT INTO users(username, password, email)
		VALUES (:username, :password, :email)
		RETURNING id`)
	if err != nil {
		return fmt.Errorf("persistence<pgUser.Create>: %w", err)
	}
	var userId uint64
	if err := stmt.Get(&userId, row); err != nil {
		return fmt.Errorf("persistence<pgUser.Create>: %w", err)
	}

	query := `
		INSERT INTO register_outbox(user_id, is_done)
		VALUES ($1, $2)`
	args := []any{userId, false}
	if _, err := tx.Exec(query, args...); err != nil {
		return fmt.Errorf("persistence<pgUser.Create>: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("persistence<pgUser.Create>: %w", err)
	}
	return nil
}

func (repo pgUser) Update(a account.User) error {
	row := newPgUserRow(a)
	query := `
		UPDATE users 
		SET
			username = :username,
			password = :password,
			email = :email
		WHERE id = :id`
	if _, err := repo.db.NamedExec(query, row); err != nil {
		return fmt.Errorf("persistence<pgUser.Update>: %w", err)
	}
	return nil
}

func (repo pgUser) GetRegisterOutbox(count uint) ([]message.UserRegistered, error) {
	rows := new([]pgRegistrationOutboxRow)
	query := `
		SELECT *
		FROM register_outbox
		WHERE is_done = false
		LIMIT $1 `
	args := []any{count}
	if err := repo.db.Select(rows, query, args...); err != nil {
		return []message.UserRegistered{}, fmt.Errorf("persistence<pgUser.GetRegisterOutbox>: %w", err)
	}

	outbox := []message.UserRegistered{}
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
		return fmt.Errorf("persistence<pgUser.ResolveRegisterOutbox>: %w", err)
	}

	if _, err = repo.db.Exec(repo.db.Rebind(query), args...); err != nil {
		return fmt.Errorf("persistence<pgUser.ResolveRegisterOutbox>: %w", err)
	}
	return nil
}
