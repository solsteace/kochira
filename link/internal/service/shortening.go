package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/link/internal/domain/shortening"
	"github.com/solsteace/kochira/link/internal/domain/shortening/store"
	"github.com/solsteace/kochira/link/internal/persistence"
)

type Shortening struct {
	store store.Link[persistence.ShorteningQueryParams]
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
	newLink.NewShortened()

	_, err = s.store.Create(newLink)
	if err != nil {
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
	if err != nil {
		return fmt.Errorf("service<Shortening.UpdateById>: %w", err)
	}
	if oldLink.UserId() != userId {
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
	if link.UserId() != userId {
		err := oops.Forbidden{
			Err: errors.New("You don't have access to this link")}
		return fmt.Errorf("service<Shortening.DeleteById>: %w", err)
	}

	if err := s.store.DeleteById(id); err != nil {
		return fmt.Errorf("service<Shortening.DeleteById>: %w", err)
	}
	return nil
}

// Activates a link, if the user is still eligible to do so
func (s Shortening) Initialize(
	id uint64,
	userId uint64,
	lifetime time.Duration,
	linkCountLimit uint,
	subscriptionExpiration time.Time,
) error {
	link, err := s.store.GetById(id)
	if err != nil {
		return fmt.Errorf("service<Shortening.Initialize>: %w", err)
	}

	linkCount, err := s.store.CountByUserId(link.UserId())
	if err != nil {
		return fmt.Errorf("service<Shortening.Initialize>: %w", err)
	}
	if linkCount > linkCountLimit {
		err := errors.New("Link quota had exceeded")
		return fmt.Errorf("service<Shortening.Initialize>: %w", err)
	}

	now := time.Now()
	linkId := link.Id()
	newLink, err := shortening.NewLink(
		&linkId,
		link.UserId(),
		link.Shortened(),
		link.Destination(),
		true,
		now,
		now.Add(lifetime))
	if err := s.store.Update(newLink); err != nil {
		return fmt.Errorf("service<Shortening.Initialize>: %w", err)
	}
	return nil
}

func NewShortening(store store.Link[persistence.ShorteningQueryParams]) Shortening {
	return Shortening{store}
}
