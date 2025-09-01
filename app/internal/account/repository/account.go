package repository

import "github.com/solsteace/kochira/internal/account/domain"

type Account interface {
	GetById(id uint) (domain.User, error)
	GetByUsername(username string) (domain.User, error)

	Create(a domain.User) error
	Update(a domain.User) error
	DeleteById(id uint) error
}
