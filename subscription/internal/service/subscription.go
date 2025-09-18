package service

import (
	"fmt"
	"slices"
	"time"

	"github.com/solsteace/kochira/subscription/internal/domain"
	domainService "github.com/solsteace/kochira/subscription/internal/domain/service"
	"github.com/solsteace/kochira/subscription/internal/repository"
)

type Subscription struct {
	repo             repository.Subscription
	subscriptionPerk domainService.SubscriptionPerks
}

func NewSubscription(
	repo repository.Subscription,
	subscriptionPerk domainService.SubscriptionPerks,
) Subscription {
	return Subscription{repo, subscriptionPerk}
}

func (ss Subscription) GetByUserId(id uint64) (domain.Subscription, error) {
	subscription, err := ss.repo.GetByOwner(id)
	if err != nil {
		return domain.Subscription{}, fmt.Errorf(
			"internal<SubscriptionService.GetByUserId>: %w", err)
	}
	return subscription, nil
}

func (ss Subscription) InferPerks(userId uint64) (time.Duration, uint) {
	subscription, err := ss.repo.GetByOwner(userId)
	if err != nil {
		return 0, 0
	}
	return ss.subscriptionPerk.Infer(subscription)
}

func (ss Subscription) Init(userId []uint64) error {
	existingSubscriptions, err := ss.repo.CheckManyByOwner(userId)
	if err != nil {
		return fmt.Errorf("internal<SubscriptionService.Init>: %w", err)
	}

	now := time.Now()
	subscriptions := []domain.Subscription{}
	for _, uId := range userId {
		// Number of new users within small time window (say, 2 - 3 seconds) won't
		// be large enough, so using linear search would be okay for now
		if slices.Contains(existingSubscriptions, uId) {
			continue
		}

		s, err := domain.NewSubscription(nil, uId, now)
		if err != nil {
			return fmt.Errorf("internal<SubscriptionService.Init>: %w", err)
		}
		subscriptions = append(subscriptions, s)
	}

	if err := ss.repo.Create(subscriptions); err != nil {
		return fmt.Errorf("internal<SubscriptionService.Init>: %w", err)
	}
	return nil
}

func (ss Subscription) Order(
	userId uint64,
	packageName string,
	quantity uint,
) {

}

func (ss Subscription) Extend(userId uint64, packageName string) {
}
