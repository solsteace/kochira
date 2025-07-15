package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solsteace/kochira/account/domain"
)

type PgAccount struct {
	Conn *pgxpool.Pool
}

func (pa PgAccount) FindMany(offset, limit int) ([]domain.Account, error) {
	return []domain.Account{}, nil
}

func (pa PgAccount) FindById(id int) (domain.Account, error) {
	row, err := pa.Conn.Query(
		context.Background(),
		"SELECT * FROM users WHERE id=1",
	)
	if err != nil {
		return domain.Account{}, nil
	}

	if found := row.Next(); !found {
		return domain.Account{}, nil
	}

	var username, email, password string
	row.Scan(nil, &username, &password, &email)
	account := [1]domain.Account{
		domain.Account{
			Id:       id,
			Username: username,
			Password: password,
			Email:    email},
	}[0]

	return account, nil
}

func (pa PgAccount) FindByEmail(email string) (domain.Account, error) {
	return domain.Account{}, nil
}

func (pa PgAccount) FindByUsername(username string) (domain.Account, error) {
	return domain.Account{}, nil
}
