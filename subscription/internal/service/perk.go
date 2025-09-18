package service

import (
	"fmt"

	domainService "github.com/solsteace/kochira/subscription/internal/domain/service"
	"github.com/solsteace/kochira/subscription/internal/repository"
)

type Perk struct {
	repo             repository.Subscription
	subscriptionPerk domainService.SubscriptionPerk
}

func (p Perk) Infer(userId uint64) error {
	subscription, err := p.repo.GetByOwner(userId)
	if err != nil {
		return fmt.Errorf("service<Perk.Infer>: %w", err)
	}

	lifetime, maxLink := p.subscriptionPerk.Infer(subscription)
	fmt.Println(lifetime)
	fmt.Println(maxLink)
	return nil
}

func NewPerk(
	repo repository.Subscription,
	subscriptionPerk domainService.SubscriptionPerk,
) Perk {
	return Perk{repo, subscriptionPerk}
}
