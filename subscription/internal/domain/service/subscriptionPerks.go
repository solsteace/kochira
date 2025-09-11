package service

import "time"

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

type SubscriptionPerks struct {
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
) SubscriptionPerks {
	return SubscriptionPerks{basic, premium, deviation}
}

func (l SubscriptionPerks) Infer(subscriptionExpiration time.Time) (time.Duration, uint) {
	diff := subscriptionExpiration.Sub(time.Now()) - l.deviation
	if diff <= 0 {
		return l.premium.expiration, l.premium.linkLimit
	}
	return l.basic.expiration, l.basic.linkLimit
}
