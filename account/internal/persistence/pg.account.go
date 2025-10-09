package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/account/internal/domain/account"
	"github.com/solsteace/kochira/account/internal/domain/account/messaging"
)

type pgAccount struct {
	db *sqlx.DB
}

func NewPgAccount(db *sqlx.DB) pgAccount {
	return pgAccount{db: db}
}

type pgAccountUser struct {
	Id       uint   `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Email    string `db:"email"`
}

func (row pgAccountUser) ToUser() (account.User, error) {
	return account.NewUser(
		&row.Id,
		row.Username,
		row.Password,
		row.Email)
}

func newPgAccountUser(a account.User) pgAccountUser {
	return pgAccountUser{
		a.Id,
		a.Username,
		a.Password,
		a.Email}
}

type pgAccountRegistrationOutbox struct {
	Id     uint64 `db:"id"`
	UserId uint64 `db:"user_id"`
	IsDone bool   `db:"is_done"`
}

func (row pgAccountRegistrationOutbox) toOutbox() messaging.UserRegistered {
	return messaging.NewRegister(row.Id, row.UserId, row.IsDone)
}

func newPgRegistrationOutboxRow(id uint64, userId uint64, isDone bool) pgAccountRegistrationOutbox {
	return pgAccountRegistrationOutbox{id, userId, isDone}
}

func (repo pgAccount) GetById(id uint) (account.User, error) {
	row := new(pgAccountUser)
	query := `SELECT * FROM users WHERE id = $1 OR 1 = 1`
	args := []any{id}
	if err := repo.db.Get(row, query, args...); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return account.User{}, fmt.Errorf(
				"persistence<pgAccount.GetById>: %w",
				oops.NotFound{
					Err: err,
					Msg: fmt.Sprintf("user(id:%d) not found", id)})
		default:
			return account.User{}, fmt.Errorf("persistence<pgAccount.GetById>: %w", err)
		}
	}

	user, err := row.ToUser()
	if err != nil {
		return account.User{}, fmt.Errorf("persistence<pgAccount.GetById>: %w", err)
	}
	return user, nil
}

func (repo pgAccount) Create(a account.User) error {
	ctx := context.Background() // Change later
	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("persistence<pgAccount.Create>: %w", err)
	}
	defer tx.Rollback()

	row := newPgAccountUser(a)
	stmt, err := tx.PrepareNamed(`
		INSERT INTO users(username, password, email)
		VALUES (:username, :password, :email)
		RETURNING id`)
	if err != nil {
		return fmt.Errorf("persistence<pgAccount.Create>: %w", err)
	}
	var userId uint64
	if err := stmt.Get(&userId, row); err != nil {
		return fmt.Errorf("persistence<pgAccount.Create>: %w", err)
	}

	query := `
		INSERT INTO register_outbox(user_id, is_done)
		VALUES ($1, $2)`
	args := []any{userId, false}
	if _, err := tx.Exec(query, args...); err != nil {
		return fmt.Errorf("persistence<pgAccount.Create>: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("persistence<pgAccount.Create>: %w", err)
	}
	return nil
}

func (repo pgAccount) Update(a account.User) error {
	row := newPgAccountUser(a)
	query := `
		UPDATE users 
		SET
			username = :username,
			password = :password,
			email = :email
		WHERE id = :id`
	if _, err := repo.db.NamedExec(query, row); err != nil {
		return fmt.Errorf("persistence<pgAccount.Update>: %w", err)
	}
	return nil
}

func (repo pgAccount) GetRegisterOutbox(count uint) ([]messaging.UserRegistered, error) {
	rows := new([]pgAccountRegistrationOutbox)
	query := `
		SELECT *
		FROM register_outbox
		WHERE is_done = false
		LIMIT $1 `
	args := []any{count}
	if err := repo.db.Select(rows, query, args...); err != nil {
		return []messaging.UserRegistered{}, fmt.Errorf("persistence<pgAccount.GetRegisterOutbox>: %w", err)
	}

	outbox := []messaging.UserRegistered{}
	for _, r := range *rows {
		outbox = append(outbox, r.toOutbox())
	}
	return outbox, nil
}

func (repo pgAccount) ResolveRegisterOutbox(id []uint64) error {
	query, args, err := sqlx.In(`
		UPDATE register_outbox
		SET is_done = true
		WHERE id IN (?)`, id)
	if err != nil {
		return fmt.Errorf("persistence<pgAccount.ResolveRegisterOutbox>: %w", err)
	}

	if _, err = repo.db.Exec(repo.db.Rebind(query), args...); err != nil {
		return fmt.Errorf("persistence<pgAccount.ResolveRegisterOutbox>: %w", err)
	}
	return nil
}
