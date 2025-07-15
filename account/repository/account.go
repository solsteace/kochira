package repository

import "github.com/solsteace/kochira/account/domain"

type AccountRepo interface {
	FindMany(offset, limit int) ([]domain.Account, error)
	FindById(id int) (domain.Account, error)
	FindByEmail(email string) (domain.Account, error)
	FindByUsername(username string) (domain.Account, error)
}
