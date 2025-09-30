package service

import (
	"fmt"
	"slices"
	"time"

	"github.com/solsteace/kochira/subscription/internal/domain/subscription"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/service"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/store"
)

const (
	CreateSubcriptionQueue = "create.subscription"
	CheckSubscriptionQueue = "check.subscription"
)

type Subscription struct {
	repo             store.Subscription
	subscriptionPerk service.PerkInferer
}

func NewSubscription(
	repo store.Subscription,
	subscriptionPerk service.PerkInferer,
) Subscription {
	return Subscription{repo, subscriptionPerk}
}

func (s Subscription) GetByUserId(id uint64) (subscription.Subscription, error) {
	sub, err := s.repo.GetByOwner(id)
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf(
			"service<Status.GetByUserId>: %w", err)
	}
	return sub, nil
}

func (s Subscription) Init(userId []uint64) error {
	existingSubscriptions, err := s.repo.CheckManyByOwner(userId)
	if err != nil {
		return fmt.Errorf("service<Status.Init>: %w", err)
	}

	now := time.Now()
	subscriptions := []subscription.Subscription{}
	for _, uId := range userId {
		// Number of new users within small time window (say, 2 - 3 seconds) won't
		// be large enough, so using linear search would be okay for now
		if slices.Contains(existingSubscriptions, uId) {
			continue
		}

		s, err := subscription.NewSubscription(nil, uId, now)
		if err != nil {
			return fmt.Errorf("service<Status.Init>: %w", err)
		}
		subscriptions = append(subscriptions, s)
	}

	if err := s.repo.Create(subscriptions); err != nil {
		return fmt.Errorf("service<Status.Init>: %w", err)
	}
	return nil
}

func (p Subscription) Infer(userId uint64) error {
	subscription, err := p.repo.GetByOwner(userId)
	if err != nil {
		return fmt.Errorf("service<Perk.Infer>: %w", err)
	}

	lifetime, maxLink := p.subscriptionPerk.Infer(subscription)
	fmt.Println(lifetime)
	fmt.Println(maxLink)
	return nil
}
