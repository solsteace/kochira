package service

import (
	"time"

	"github.com/solsteace/kochira/subscription/internal/domain"
)

type perks struct {
	expiration time.Duration
	linkLimit  uint
}

func NewPerks(
	expiration time.Duration,
	linkLimit uint,
) perks {
	return perks{expiration, linkLimit}
}

type SubscriptionPerk struct {
	basic   perks
	premium perks

	// How early subscriptions considered as expired from actual expiration time.
	// It prevents premium perks being used just before or even during the
	// premium-to-basic perks conversion.
	deviation time.Duration
}

func NewSubscriptionPerks(
	basic perks,
	premium perks,
	deviation time.Duration,
) SubscriptionPerk {
	return SubscriptionPerk{basic, premium, deviation}
}

func (l SubscriptionPerk) Infer(subscription domain.Subscription) (time.Duration, uint) {
	expiration := subscription.ExpiredAt()
	diff := expiration.Sub(time.Now()) - l.deviation
	if diff <= 0 {
		return l.premium.expiration, l.premium.linkLimit
	}
	return l.basic.expiration, l.basic.linkLimit
}
