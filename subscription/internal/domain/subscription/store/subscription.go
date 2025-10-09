package store

import (
	"github.com/solsteace/kochira/subscription/internal/domain/subscription"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/messaging"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/value"
)

type Subscription interface {
	GetByOwner(id uint64) (subscription.Subscription, error)
	Create(s []subscription.Subscription) error // Creates new subscription while ignoring the existing ones

	// Events ===========

	CreateSubscriptionChecked(contextId uint64, usecase string, perk value.Perk) error
	GetSubscriptionChecked(limit uint) ([]messaging.SubscriptionChecked, error)
	ResolveSubscriptionChecked(id []uint64) error

	WatchExpiringSubscription(limit uint) error
	GetSubscriptionExpired(limit uint) ([]messaging.SubscriptionExpired, error)
	ResolveSubscriptionExpired(id []uint64) error
}
