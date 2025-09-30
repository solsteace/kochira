package store

import "github.com/solsteace/kochira/subscription/internal/domain/subscription"

type Subscription interface {
	// Query
	GetByOwner(id uint64) (subscription.Subscription, error)
	CheckManyByOwner(id []uint64) ([]uint64, error)

	// Command
	Create(s []subscription.Subscription) error
	Update(s subscription.Subscription) error
	Delete(userId uint64) error
}
