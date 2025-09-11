package repository

import "github.com/solsteace/kochira/subscription/internal/domain"

type Subscription interface {
	// Query
	GetByOwner(id uint64) (domain.Subscription, error)

	// Command
	Create(s []domain.Subscription) error
	Update(s domain.Subscription) error
	Delete(userId uint64) error
}
