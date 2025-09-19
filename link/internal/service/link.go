package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/link/internal/domain"
	"github.com/solsteace/kochira/link/internal/repository"
	"github.com/solsteace/kochira/link/internal/view"
)

type Link struct {
	repo repository.Link
}

func NewLink(domainRepo repository.Link) Link {
	return Link{domainRepo}
}

// CRUD get many
func (ls Link) GetSelf(userId uint64, page, limit *uint) ([]view.Link, error) {
	qParams := repository.NewLinkQueryParams(page, limit)
	links, err := ls.repo.GetManyByUser(userId, qParams)
	if err != nil {
		return []view.Link{}, fmt.Errorf("service<Link.GetSelf>: %w", err)
	}
	return links, nil
}

// CRUD get one
func (ls Link) GetById(userId, id uint64) (
	view.Link, error,
) {
	link, err := ls.repo.GetById(id)
	if err != nil {
		return view.Link{}, fmt.Errorf("service<Link.GetById>: %w", err)
	}
	return link, nil
}

// CRUD create
func (ls Link) Create(userId uint64, destination string) error {
	now := time.Now()
	newLink, err := domain.NewLink(
		nil,
		userId,
		"",
		destination,
		true,
		now,
		now)
	if err != nil {
		return fmt.Errorf("service<Link.Create>: %w", err)
	}
	newLink.NewShortened()

	_, err = ls.repo.Create(newLink)
	if err != nil {
		return fmt.Errorf("service<Link.Create>: %w", err)
	}

	// Publish "check.subscription" command

	return nil
}

// CRUD update
func (ls Link) UpdateById(
	userId uint64,
	id uint64,
	shortened string,
	destination string,
	IsOpen bool,
) error {
	oldLink, err := ls.repo.Load(id)
	if err != nil {
		return fmt.Errorf("service<Link.UpdateById>: %w", err)
	}
	if oldLink.UserId() != userId {
		err := oops.Forbidden{
			Err: errors.New("You don't have access to this link")}
		return fmt.Errorf("service<Link.UpdateById>: %w", err)
	}

	newLink, err := domain.NewLink(
		&id,
		userId,
		oldLink.Shortened(), // Updated after subscription check
		destination,
		oldLink.IsOpen(), // Updated after subscription check
		oldLink.UpdatedAt(),
		oldLink.ExpiredAt())
	if err != nil {
		return fmt.Errorf("service<Link.UpdateById>: %w", err)
	}

	if err := ls.repo.Update(newLink); err != nil {
		return fmt.Errorf("service<Link.UpdateById>: %w", err)
	}

	// Publish "check.subscription" command

	return nil
}

// CRUD delete
func (ls Link) DeleteById(userId, id uint64) error {
	link, err := ls.repo.GetById(id)
	if err != nil {
		return fmt.Errorf("service<Link.DeleteById>: %w", err)
	}
	if link.UserId != userId {
		err := oops.Forbidden{
			Err: errors.New("You don't have access to this link")}
		return fmt.Errorf("service<Link.DeleteById>: %w", err)
	}

	if err := ls.repo.DeleteById(id); err != nil {
		return fmt.Errorf("service<Link.DeleteById>: %w", err)
	}
	return nil
}

// Redirects the user to the destination based on given shortened URI
func (ls Link) Redirect(shortened string) (string, error) {
	link, err := ls.repo.GetByShortened(shortened)
	if err != nil {
		return "", fmt.Errorf("service<Link.Redirect>: %w", err)
	}
	switch {
	case !link.IsOpen:
		err := oops.Forbidden{
			Err: errors.New("This link is not opened by the owner"),
			Msg: "This link is not opened by the owner"}
		return "", fmt.Errorf("service<Link.Redirect>: %w", err)
	case time.Now().Sub(link.ExpiredAt) > 0:
		err := oops.Forbidden{
			Err: errors.New("This link had already expired"),
			Msg: "This link had already expired"}
		return "", fmt.Errorf("service<Link.Redirect>: %w", err)
	}
	return link.Destination, nil
}

// Activates a link, if the user is still eligible to do so
func (ls Link) Initialize(
	id uint64,
	userId uint64,
	lifetime time.Duration,
	linkCountLimit uint,
	subscriptionExpiration time.Time,
) error {
	link, err := ls.repo.GetById(id)
	if err != nil {
		return fmt.Errorf("service<Link.Initialize>: %w", err)
	}

	linkCount, err := ls.repo.CountByUserId(link.UserId)
	if err != nil {
		return fmt.Errorf("service<Link.Initialize>: %w", err)
	}
	if linkCount > linkCountLimit {
		err := errors.New("Link quota had exceeded")
		return fmt.Errorf("service<Link.Initialize>: %w", err)
	}

	now := time.Now()
	linkId := link.Id
	newLink, err := domain.NewLink(
		&linkId,
		link.UserId,
		link.Shortened,
		link.Destination,
		true,
		now,
		now.Add(lifetime))
	if err := ls.repo.Update(newLink); err != nil {
		return fmt.Errorf("service<Link.Initialize>: %w", err)
	}
	return nil
}
