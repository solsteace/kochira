package repository

import "github.com/solsteace/kochira/account/internal/domain"

type User interface {
	GetById(id uint) (domain.User, error)
	GetByUsername(username string) (domain.User, error)

	Create(a domain.User) error
	Update(a domain.User) error
	DeleteById(id uint) error
}
