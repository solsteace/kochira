package service

import (
	"errors"
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
	CheckSubscriptionQueue = "check.subscription"
	FinishShorteningQueue  = "finish.shortening"
)

type Shortening struct {
	store     store.Link[persistence.ShorteningQueryParams]
	messenger *utility.Amqp // interface later
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
		destination,
		true,
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
	shortened string,
	destination string,
	IsOpen bool,
) error {
	oldLink, err := s.store.GetById(id)
	switch {
	case err != nil:
		return fmt.Errorf("service<Shortening.UpdateById>: %w", err)
	case !oldLink.AccessibleBy(userId):
		err := oops.Forbidden{
			Err: errors.New("You don't have access to this link")}
		return fmt.Errorf("service<Shortening.UpdateById>: %w", err)
	}

	newLink, err := shortening.NewLink(
		&id,
		userId,
		oldLink.Shortened(), // Updated after subscription check
		destination,
		oldLink.IsOpen(), // Updated after subscription check
		oldLink.UpdatedAt(),
		oldLink.ExpiredAt())
	if err != nil {
		return fmt.Errorf("service<Shortening.UpdateById>: %w", err)
	}

	if err := s.store.Update(newLink); err != nil {
		return fmt.Errorf("service<Shortening.UpdateById>: %w", err)
	}
	return nil
}

// CRUD delete
func (s Shortening) DeleteById(userId, id uint64) error {
	link, err := s.store.GetById(id)
	if err != nil {
		return fmt.Errorf("service<Shortening.DeleteById>: %w", err)
	}
	if !link.AccessibleBy(userId) {
		err := oops.Forbidden{
			Err: errors.New("You don't have access to this link")}
		return fmt.Errorf("service<Shortening.DeleteById>: %w", err)
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
	}
	if len(msg) == 0 {
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
	}
	if len(msg) == 0 {
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
		err := oops.Forbidden{
			Err: errors.New("You don't have access to this link")}
		return fmt.Errorf("service<Shortening.HandleLinkShortened>: %w", err)
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
		err := errors.New("Quota for simultaneous active shortened links had ran out")
		return fmt.Errorf("service<Shortening.HandleLinkShortened>: %w", err)
	}

	now := time.Now()
	oldLinkId := oldLink.Id()
	newLink, err := shortening.NewLink(
		&oldLinkId,
		oldLink.UserId(),
		oldLink.Shortened(),
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
// (A copy from HandleLinkShortened, handle if the handler needs to fetch user stats).
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
func (s Shortening) HandleShortConfigured(
	msgId uint64,
	allowEditShortUrl bool,
) error {
	msgCtx, err := s.store.GetShortConfiguredById(msgId)
	if err != nil {
		return fmt.Errorf("service<Shortening.HandleShortConfigured>: %w", err)
	}

	oldLink, err := s.store.GetById(msgCtx.LinkId())
	if err != nil {
		return fmt.Errorf("service<Shortening.HandleShortConfigured>: %w", err)
	} else if !oldLink.AccessibleBy(msgCtx.UserId()) {
		err := oops.Forbidden{
			Err: errors.New("You don't have access to this link")}
		return fmt.Errorf("service<Shortening.HandleShortConfigured>: %w", err)
	}

	if oldLink.ShortChanged(msgCtx.Shortened()) && !allowEditShortUrl {
		err := oops.Forbidden{
			Err: errors.New("Your subscription doesn't allow short editing")}
		return fmt.Errorf("service<Shortening.HandleShortConfigured>: %w", err)
	}

	linkId := oldLink.Id()
	newLink, err := shortening.NewLink(
		&linkId,
		oldLink.UserId(),
		msgCtx.Shortened(),
		msgCtx.Destination(),
		oldLink.IsOpen(),
		time.Now(),
		oldLink.ExpiredAt())
	if err != nil {
		return fmt.Errorf("service<Shortening.HandleShortConfigured>: %w", err)
	}

	if err := s.store.Update(newLink); err != nil {
		return fmt.Errorf("service<Shortening.HandleLinkShortened>: %w", err)
	}
	return nil
}

func NewShortening(
	store store.Link[persistence.ShorteningQueryParams],
	messenger *utility.Amqp,
) Shortening {
	return Shortening{store, messenger}
}
