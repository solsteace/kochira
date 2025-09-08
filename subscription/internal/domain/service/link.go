package domain

import "time"

type subscriptionPerks struct {
	expiration time.Duration
	linkLimit  uint
}

func NewSubscriptionPerks(
	expiration time.Duration,
	linkLimit uint,
) subscriptionPerks {
	return subscriptionPerks{expiration, linkLimit}
}

type LinkSubscription struct {
	basic   subscriptionPerks
	premium subscriptionPerks

	// How early subscriptions considered expired from actual expiration time.
	// It prevents premium perks being used just before or even during the
	// premium-to-basic perks conversion.
	deviation time.Duration
}

func NewLink(
	basic subscriptionPerks,
	premium subscriptionPerks,
	deviation time.Duration,
) LinkSubscription {
	return LinkSubscription{basic, premium, deviation}
}

func (l LinkSubscription) InferPerks(subscriptionExpiration time.Time) (time.Duration, uint) {
	diff := subscriptionExpiration.Sub(time.Now()) - l.deviation
	if diff <= 0 {
		return l.premium.expiration, l.premium.linkLimit
	}
	return l.basic.expiration, l.basic.linkLimit
}
