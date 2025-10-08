package store

import (
	"github.com/solsteace/kochira/subscription/internal/domain/subscription"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/messaging"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/value"
)

type Subscription interface {
	GetByOwner(id uint64) (subscription.Subscription, error)
	FilterExisting(id []uint64) ([]uint64, error)
	Create(s []subscription.Subscription) error

	// Events ===========

	CreateSubscriptionChecked(contextId uint64, usecase string, perk value.Perk) error
	GetSubscriptionChecked(limit uint) ([]messaging.SubscriptionChecked, error)
	ResolveSubscriptionChecked(id []uint64) error

	WatchExpiringSubscription(limit uint) error
	GetSubscriptionExpired(limit uint) ([]messaging.SubscriptionExpired, error)
	ResolveSubscriptionExpired(id []uint64) error
}
