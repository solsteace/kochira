package repository

import (
	"github.com/solsteace/kochira/account/internal/domain"
	"github.com/solsteace/kochira/account/internal/domain/outbox"
)

type User interface {
	GetById(id uint) (domain.User, error)
	GetByUsername(username string) (domain.User, error)

	Create(a domain.User) error
	Update(a domain.User) error

	// Fetches pending `registration` outboxes
	GetRegisterOutbox(count uint) ([]outbox.Register, error)

	// Resolves pending `registration` outboxes
	ResolveRegisterOutbox(id []uint64) error
}
