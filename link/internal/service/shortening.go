package service

import (
	"fmt"
	"time"

	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/link/internal/domain/shortening"
	shorteningMessaging "github.com/solsteace/kochira/link/internal/domain/shortening/messaging"
	"github.com/solsteace/kochira/link/internal/domain/shortening/store"
	"github.com/solsteace/kochira/link/internal/persistence"
	"github.com/solsteace/kochira/link/internal/utility"
)

const (
	CheckSubscriptionQueue   = "check.subscription"
	FinishShorteningQueue    = "finish.shortening"
	SubscriptionExpiredQueue = "test"
)

type Shortening struct {
	store     store.Link[persistence.ShorteningQueryParams]
	messenger *utility.Amqp // interface later
}

func NewShortening(
	store store.Link[persistence.ShorteningQueryParams],
	messenger *utility.Amqp,
) Shortening {
	return Shortening{store, messenger}
}

func (s Shortening) GetSelf(userId uint64, page, limit *uint) ([]shortening.Link, error) {
	qParams := persistence.NewShorteningQueryParams(page, limit)
	links, err := s.store.GetManyByUser(userId, qParams)
	if err != nil {
		return []shortening.Link{}, fmt.Errorf("service<Shortening.GetSelf>: %w", err)
	}
	return links, nil
}

func (s Shortening) GetById(userId, id uint64) (shortening.Link, error) {
	link, err := s.store.GetById(id)
	if err != nil {
		return shortening.Link{}, fmt.Errorf("service<Shortening.GetById>: %w", err)
	}
	return link, nil
}

func (s Shortening) Create(userId uint64, destination string) error {
	now := time.Now()
	newLink, err := shortening.NewLink(
		nil,
		userId,
		"",
		"",
		destination,
		false,
		now,
		now)
	if err != nil {
		return fmt.Errorf("service<Shortening.Create>: %w", err)
	}

	newLink.Shorten()
	if err := s.store.Create(newLink); err != nil {
		return fmt.Errorf("service<Shortening.Create>: %w", err)
	}
	return nil
}

func (s Shortening) UpdateById(
	userId uint64,
	id uint64,
	alias string,
	destination string,
	isOpen bool,
) error {
	oldLink, err := s.store.GetById(id)
	if err != nil {
		return fmt.Errorf("service<Shortening.UpdateById>: %w", err)
	} else if !oldLink.AccessibleBy(userId) {
		return fmt.Errorf(
			"service<Shortening.DeleteById>: %w",
			oops.Forbidden{Msg: "You don't have access to this link"})
	}

	requirePremiumSubscription := oldLink.HasCustomAlias()
	newLink, err := shortening.NewLink(
		&id,
		userId,
		oldLink.Shortened(),
		alias,
		destination,
		isOpen,
		oldLink.UpdatedAt(),
		oldLink.ExpiredAt())
	if err != nil {
		return fmt.Errorf("service<Shortening.UpdateById>: %w", err)
	}

	if requirePremiumSubscription {
		err = s.store.UpdateWithSubscription(newLink)
	} else {
		err = s.store.Update(newLink)
	}
	if err != nil {
		return fmt.Errorf("service<Shortening.UpdateById>: %w", err)
	}
	return nil
}

// CRUD delete
func (s Shortening) DeleteById(userId, id uint64) error {
	link, err := s.store.GetById(id)
	if err != nil {
		return fmt.Errorf("service<Shortening.DeleteById>: %w", err)
	} else if !link.AccessibleBy(userId) {
		return fmt.Errorf(
			"service<Shortening.DeleteById>: %w",
			oops.Forbidden{Msg: "You don't have access to this link"})
	}

	if err := s.store.DeleteById(id); err != nil {
		return fmt.Errorf("service<Shortening.DeleteById>: %w", err)
	}
	return nil
}

// ===================================
// Events
// ===================================

func (s Shortening) PublishLinkShortened(
	maxMsg uint,
	serialize func(msg shorteningMessaging.LinkShortened) ([]byte, error),
) error {
	msg, err := s.store.GetLinkShortened(maxMsg)
	if err != nil {
		return fmt.Errorf("service<Shortening.PublishLinkShortened>: %w", err)
	} else if len(msg) == 0 {
		return nil
	}

	resolved := []uint64{}
	for _, m := range msg {
		payload, err := serialize(m)
		if err != nil {
			return fmt.Errorf("service<Shortening.PublishLinkShortened>: %w", err)
		}

		err = s.messenger.Publish("default", payload, utility.NewDefaultAmqpPublishOpts(
			"", CheckSubscriptionQueue, "application/json"))
		if err != nil {
			return fmt.Errorf("service<Shortening.PublishLinkShortened>: %w", err)
		}
		resolved = append(resolved, m.Id())
	}

	if err := s.store.ResolveLinkShortened(resolved); err != nil {
		return fmt.Errorf("service<Shortening.PublishLinkShortened>: %w", err)
	}
	return nil
}

func (s Shortening) PublishShortConfigured(
	maxMsg uint,
	serialize func(msg shorteningMessaging.ShortConfigured) ([]byte, error),
) error {
	msg, err := s.store.GetShortConfigured(maxMsg)
	if err != nil {
		return fmt.Errorf("service<Shortening.PublishShortConfigured>: %w", err)
	} else if len(msg) == 0 {
		return nil
	}

	resolved := []uint64{}
	for _, m := range msg {
		payload, err := serialize(m)
		if err != nil {
			return fmt.Errorf("service<Shortening.PublishShortConfigured>: %w", err)
		}

		opts := utility.NewDefaultAmqpPublishOpts("", CheckSubscriptionQueue, "application/json")
		if err = s.messenger.Publish("default", payload, opts); err != nil {
			return fmt.Errorf("service<Shortening.PublishShortConfigured>: %w", err)
		}
		resolved = append(resolved, m.Id())
	}

	if err := s.store.ResolveShortConfigured(resolved); err != nil {
		return fmt.Errorf("service<Shortening.PublishShortConfigured>: %w", err)
	}
	return nil
}

// # TODO
//
// There's a chance of race condition if this function was called in many goroutines
// with the same userId.
//
// ## Reason
//
// Race condition "happens in the database-level". There's a chance that, after goroutine
// `A` fetches current user stats, goroutine `B“ do writes that affects this stats. This
// would disrupt the subscription perk handling.
//
// For example, goroutine `A` thought user `X` only has 1 link quota left. However,
// after `A` deducted this, goroutine `B` adds a new link which reduces the quota
// to 0. This makes `A`'s view on current user stats doesn't sync with what actually
// in the database.
//
// ## Suggestions
//
// - optimistic locking (via updatedAt). Challenge: Requires retry mechanism whenever needed
//
// - row locking. Challenge: May need to inject business logic between database queries
func (s Shortening) HandleLinkShortened(
	msgId uint64,
	lifetime time.Duration,
	linkCountLimit uint,
) error {
	msgCtx, err := s.store.GetLinkShortenedById(msgId)
	if err != nil {
		return fmt.Errorf("service<Shortening.HandleLinkShortened>: %w", err)
	}

	oldLink, err := s.store.GetById(msgCtx.LinkId())
	if err != nil {
		return fmt.Errorf("service<Shortening.HandleLinkShortened>: %w", err)
	} else if !oldLink.AccessibleBy(msgCtx.UserId()) {
		return fmt.Errorf(
			"service<Shortening.HandleLinkShortened>: %w",
			oops.Forbidden{Msg: fmt.Sprintf(
				"User(id:%d) doesn't have access to Link(id:%d)",
				msgCtx.UserId(), msgCtx.LinkId())})
	}

	// Skip the-not-new-links. By logic, these links would always
	// have their `expiredAt` < CURRENT_TIMESTAMP. The logic also works as
	// the "idempotency token"
	if time.Now().Sub(oldLink.ExpiredAt()) < 0 {
		return nil
	}

	stats, err := s.store.CountByUserIdExcept(msgCtx.UserId(), msgCtx.LinkId())
	if err != nil {
		return fmt.Errorf("service<Shortening.HandleLinkShortened>: %w", err)
	} else if !stats.HasQuota(linkCountLimit, 1) {
		return fmt.Errorf(
			"service<Shortening.HandleLinkShortened>: %w",
			oops.Forbidden{Msg: fmt.Sprintf(
				"Quota for simultaneous active shortened links had ran out (limit: %d; have: %d)",
				linkCountLimit, stats.ActiveLinks())})
	}

	now := time.Now()
	oldLinkId := oldLink.Id()
	newLink, err := shortening.NewLink(
		&oldLinkId,
		oldLink.UserId(),
		oldLink.Shortened(),
		oldLink.Alias(),
		oldLink.Destination(),
		true,
		now,
		now.Add(lifetime))
	if err != nil {
		return fmt.Errorf("service<Shortening.HandleLinkShortened>: %w", err)
	}

	if err := s.store.Update(newLink); err != nil {
		return fmt.Errorf("service<Shortening.HandleLinkShortened>: %w", err)
	}
	return nil
}

// # TODO
//
// (Read `HandleLinkShortened` and handle if the handler needs to fetch user stats).
func (ss Shortening) HandleShortConfigured(
	msgId uint64,
	allowEditShortUrl bool,
) error {
	msgCtx, err := ss.store.GetShortConfiguredById(msgId)
	if err != nil {
		return fmt.Errorf("service<Shortening.HandleShortConfigured>: %w", err)
	}

	oldLink, err := ss.store.GetById(msgCtx.LinkId())
	if err != nil {
		return fmt.Errorf("service<Shortening.HandleShortConfigured>: %w", err)
	} else if !oldLink.AccessibleBy(msgCtx.UserId()) {
		return fmt.Errorf(
			"service<Shortening.HandleShortConfigured>: %w",
			oops.Forbidden{Msg: "You don't have access to this link"})
	}

	if oldLink.HasCustomAlias() && !allowEditShortUrl {
		return fmt.Errorf(
			"service<Shortening.HandleShortConfigured>: %w",
			oops.Forbidden{Msg: "Your subscription doesn't allow short editing"})
	}

	linkId := oldLink.Id()
	newLink, err := shortening.NewLink(
		&linkId,
		oldLink.UserId(),
		oldLink.Shortened(),
		msgCtx.Alias(),
		msgCtx.Destination(),
		msgCtx.IsOpen(),
		time.Now(),
		oldLink.ExpiredAt())
	if err != nil {
		return fmt.Errorf("service<Shortening.HandleShortConfigured>: %w", err)
	}

	if err := ss.store.Update(newLink); err != nil {
		return fmt.Errorf("service<Shortening.HandleLinkShortened>: %w", err)
	}
	return nil
}

func (ss Shortening) HandleSubscriptionExpired(
	userId uint64,
	linkCountLimit uint,
	allowEditShortUrl bool,
) error {
	links, err := ss.store.GetOpenedFromOldestByUser(userId)
	if err != nil {
		return fmt.Errorf("service<Shortening.HandleSubscriptionExpired>: %w", err)
	}

	deactivatedLinks := []uint64{}
	keptLinks := []uint64{}
	for _, l := range links {
		isPremiumLink := l.HasCustomAlias()
		if isPremiumLink {
			deactivatedLinks = append(deactivatedLinks, l.Id())
		} else {
			keptLinks = append(keptLinks, l.Id())
		}
	}
	if remainingQuota := int(linkCountLimit) - len(keptLinks); remainingQuota < 0 {
		excessIdx := -1 * remainingQuota
		deactivatedLinks = append(deactivatedLinks, keptLinks[:excessIdx]...)
		keptLinks = keptLinks[excessIdx:]
	}
	if len(deactivatedLinks) == 0 {
		return nil
	}

	if err = ss.store.ApplySubscriptionExpiration(deactivatedLinks); err != nil {
		return fmt.Errorf("service<Shortening.HandleSubscriptionExpired>: %w", err)
	}
	return nil
}
