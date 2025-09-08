package internal

import (
	"errors"
	"time"

	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/link/internal/domain"
	"github.com/solsteace/kochira/link/internal/repository"
	"github.com/solsteace/kochira/link/internal/view"
)

type linkService struct {
	repo repository.Link
}

func NewLinkService(
	domainRepo repository.Link,
) linkService {
	return linkService{domainRepo}
}

// CRUD get many
func (ls linkService) GetSelf(userId uint64, page, limit *uint) (
	[]view.Link, error,
) {
	qParams := repository.NewLinkQueryParams(page, limit)
	links, err := ls.repo.GetManyByUser(userId, qParams)
	if err != nil {
		return []view.Link{}, err
	}
	return links, nil
}

// CRUD get one
func (ls linkService) GetById(userId, id uint64) (
	view.Link, error,
) {
	link, err := ls.repo.GetById(id)
	if err != nil {
		return view.Link{}, err
	}
	return link, nil
}

// CRUD create
func (ls linkService) Create(userId uint64, destination string) error {
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
		return err
	}
	newLink.NewShortened()

	_, err = ls.repo.Create(newLink)
	if err != nil {
		return err
	}

	// Publish "check.subscription" command

	return nil
}

// CRUD update
func (ls linkService) UpdateById(
	userId uint64,
	id uint64,
	shortened string,
	destination string,
	IsOpen bool,
) error {
	oldLink, err := ls.repo.Load(id)
	if err != nil {
		return err
	}
	if oldLink.UserId() != userId {
		return oops.Forbidden{
			Err: errors.New("You don't have access to this link")}
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
		return err
	}

	if err := ls.repo.Update(newLink); err != nil {
		return err
	}

	// Publish "check.subscription" command

	return nil
}

// CRUD delete
func (ls linkService) DeleteById(userId, id uint64) error {
	link, err := ls.repo.GetById(id)
	if err != nil {
		return err
	}
	if link.UserId != userId {
		return oops.Forbidden{
			Err: errors.New("You don't have access to this link")}
	}

	return ls.repo.DeleteById(id)
}

// Redirects the user to the destination based on given shortened URI
func (ls linkService) Redirect(shortened string) (string, error) {
	link, err := ls.repo.FindRedirection(shortened)
	if err != nil {
		return "", err
	}
	switch {
	case link.IsOpen == false:
		return "", oops.Forbidden{
			Err: errors.New("This link is not opened by the owner"),
			Msg: "This link is not opened by the owner"}
	case time.Now().Sub(link.ExpiredAt) > 0:
		return "", oops.Forbidden{
			Err: errors.New("This link had already expired"),
			Msg: "This link had already expired"}
	}
	return link.Destination, nil
}

// Activates a link, if the user is still eligible to do so
func (ls linkService) Initialize(
	id uint64,
	userId uint64,
	lifetime time.Duration,
	linkCountLimit uint,
	subscriptionExpiration time.Time,
) error {
	link, err := ls.repo.GetById(id)
	if err != nil {
		return err
	}

	linkCount, err := ls.repo.CountByUserId(link.UserId)
	if err != nil {
		return err
	}
	if linkCount > linkCountLimit {
		return errors.New("Link quota had exceeded")
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
		return err
	}
	return nil
}
