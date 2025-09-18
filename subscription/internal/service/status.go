package service

import (
	"fmt"
	"slices"
	"time"

	"github.com/solsteace/kochira/subscription/internal/domain"
	domainService "github.com/solsteace/kochira/subscription/internal/domain/service"
	"github.com/solsteace/kochira/subscription/internal/repository"
)

type Status struct {
	repo             repository.Subscription
	subscriptionPerk domainService.SubscriptionPerk
}

func NewStatus(
	repo repository.Subscription,
	subscriptionPerk domainService.SubscriptionPerk,
) Status {
	return Status{repo, subscriptionPerk}
}

func (s Status) GetByUserId(id uint64) (domain.Subscription, error) {
	subscription, err := s.repo.GetByOwner(id)
	if err != nil {
		return domain.Subscription{}, fmt.Errorf(
			"service<Status.GetByUserId>: %w", err)
	}
	return subscription, nil
}

func (s Status) Init(userId []uint64) error {
	existingSubscriptions, err := s.repo.CheckManyByOwner(userId)
	if err != nil {
		return fmt.Errorf("service<Status.Init>: %w", err)
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
			return fmt.Errorf("service<Status.Init>: %w", err)
		}
		subscriptions = append(subscriptions, s)
	}

	if err := s.repo.Create(subscriptions); err != nil {
		return fmt.Errorf("service<Status.Init>: %w", err)
	}
	return nil
}

func (s Status) Order(
	userId uint64,
	packageName string,
	quantity uint,
) {

}

func (ss Status) Extend(userId uint64, packageName string) {
}
