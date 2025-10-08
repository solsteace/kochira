package service

import (
	"time"

	"github.com/solsteace/kochira/subscription/internal/domain/subscription"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/value"
)

type PerkInferer struct {
	basic   value.Perk
	premium value.Perk

	// How early subscriptions considered as expired from actual expiration time.
	// It prevents premium perks being used just before or even during the
	// premium-to-basic perks conversion.
	deviation time.Duration
}

func NewPerkInferer(
	basic value.Perk,
	premium value.Perk,
	deviation time.Duration,
) PerkInferer {
	return PerkInferer{basic, premium, deviation}
}

func (l PerkInferer) Infer(subscription subscription.Subscription) value.Perk {
	expiration := subscription.ExpiredAt()
	if diff := expiration.Sub(time.Now()) - l.deviation; diff > 0 {
		return l.premium
	}
	return l.basic
}

func (l PerkInferer) Basic() value.Perk   { return l.basic }
func (l PerkInferer) Premium() value.Perk { return l.premium }
