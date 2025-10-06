package store

import (
	"github.com/solsteace/kochira/account/internal/domain/account"
	"github.com/solsteace/kochira/account/internal/domain/auth/message"
)

type User interface {
	GetById(id uint) (account.User, error)
	Create(a account.User) error
	Update(a account.User) error

	// Fetches pending `registration` outboxes
	GetRegisterOutbox(count uint) ([]message.UserRegistered, error)

	// Resolves pending `registration` outboxes
	ResolveRegisterOutbox(id []uint64) error
}
