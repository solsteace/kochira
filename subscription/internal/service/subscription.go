package service

import (
	"fmt"
	"time"

	"github.com/solsteace/kochira/subscription/internal/domain/subscription"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/messaging"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/service"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/store"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/value"
	"github.com/solsteace/kochira/subscription/internal/utility"
)

const (
	CreateSubcriptionQueue = "create.subscription"
	CheckSubscriptionQueue = "check.subscription"
	FinishShorteningQueue  = "finish.shortening"
)

type Subscription struct {
	store       store.Subscription
	perkInferer service.PerkInferer
	messenger   *utility.Amqp // Change to interface later
}

func NewSubscription(
	repo store.Subscription,
	subscriptionPerk service.PerkInferer,
	messenger *utility.Amqp,
) Subscription {
	return Subscription{repo, subscriptionPerk, messenger}
}

func (s Subscription) GetByUserId(id uint64) (subscription.Subscription, error) {
	sub, err := s.store.GetByOwner(id)
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf(
			"service<Subscription.GetByUserId>: %w", err)
	}
	return sub, nil
}

func (s Subscription) Init(userId []uint64) error {
	newSubscriptions, err := s.store.FilterExisting(userId)
	if err != nil {
		return fmt.Errorf("service<Subscription.Init>: %w", err)
	}

	now := time.Now()
	subscriptions := []subscription.Subscription{}
	for _, uId := range newSubscriptions {
		s, err := subscription.NewSubscription(nil, uId, now)
		if err != nil {
			return fmt.Errorf("service<Subscription.Init>: %w", err)
		}
		subscriptions = append(subscriptions, s)
	}

	if err := s.store.Create(subscriptions); err != nil {
		return fmt.Errorf("service<Subscription.Init>: %w", err)
	}
	return nil
}

func (p Subscription) Check(
	userId uint64,
	contextId uint64,
	usecase string,
) error {
	subscription, err := p.store.GetByOwner(userId)
	if err != nil {
		return fmt.Errorf("service<Subscription.Check>: %w", err)
	}

	perk := p.perkInferer.Infer(subscription)
	err = p.store.CreateSubscriptionChecked(contextId, usecase, perk)
	if err != nil {
		return fmt.Errorf("service<Subscription.Check>: %w", err)
	}
	return nil
}

// ==============================
// Messages
// ==============================

func (p Subscription) WatchExpiringSubscription(limit uint) error {
	if err := p.store.WatchExpiringSubscription(limit); err != nil {
		return fmt.Errorf("service<Subscription.WatchExpiringSubscription>: %w", err)
	}
	return nil
}

func (p Subscription) PublishFinishShortening(
	limit uint,
	serialize func(msg messaging.SubscriptionChecked) ([]byte, error),
) error {
	msg, err := p.store.GetSubscriptionChecked(limit)
	if err != nil {
		return fmt.Errorf("service<Subscription.PublishFinishShortening>: %w", err)
	} else if len(msg) == 0 {
		return nil
	}

	resolved := []uint64{}
	for _, m := range msg {
		payload, err := serialize(m)
		if err != nil {
			return fmt.Errorf("service<Shortening.PublishFinishShortening>: %w", err)
		}

		err = p.messenger.Publish("default", payload, utility.NewDefaultAmqpPublishOpts(
			"", FinishShorteningQueue, "application/json"))
		if err != nil {
			return fmt.Errorf("service<Shortening.PublishFinishShortening>: %w", err)
		}
		resolved = append(resolved, m.Id())
	}

	if err := p.store.ResolveSubscriptionChecked(resolved); err != nil {
		return fmt.Errorf("service<Shortening.PublishFinishShortening>: %w", err)
	}
	return nil
}

func (p Subscription) PublishSubscriptionExpired(
	limit uint,
	serialize func(msg messaging.SubscriptionExpired, perk value.Perk) ([]byte, error),
) error {
	msg, err := p.store.GetSubscriptionExpired(limit)
	if err != nil {
		return fmt.Errorf("service<Subscription.PublishSubscriptionExpired>: %w", err)
	} else if len(msg) == 0 {
		return nil
	}

	perk := p.perkInferer.Basic()
	resolved := []uint64{}
	for _, m := range msg {
		payload, err := serialize(m, perk)
		if err != nil {
			return fmt.Errorf("service<Shortening.PublishSubscriptionExpired>: %w", err)
		}

		err = p.messenger.Publish("default", payload, utility.NewDefaultAmqpPublishOpts(
			"example", "test", "application/json"))
		if err != nil {
			return fmt.Errorf("service<Shortening.PublishFinishShortening>: %w", err)
		}
		resolved = append(resolved, m.Id())
	}

	if err := p.store.ResolveSubscriptionExpired(resolved); err != nil {
		return fmt.Errorf("service<Shortening.PublishFinishShortening>: %w", err)
	}
	return nil
}
