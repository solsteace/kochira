package repository

import "github.com/solsteace/kochira/subscription/internal/domain"

type Subscription interface {
	// Query
	GetByOwner(id uint64) (domain.Subscription, error)
	CheckManyByOwner(id []uint64) ([]uint64, error)

	// Command
	Create(s []domain.Subscription) error
	Update(s domain.Subscription) error
	Delete(userId uint64) error
}
