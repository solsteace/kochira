package internal

import (
	"errors"
	"time"

	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/link/internal/domain"
	"github.com/solsteace/kochira/link/internal/repository"
)

type linkService struct {
	repo repository.Link
}

func NewLinkService(
	repo repository.Link,
) linkService {
	return linkService{repo}
}

// CRUD get many
func (ls linkService) GetSelf(userId, page, limit uint) (
	*struct{ Link []domain.Link },
	error,
) {
	result := new(struct{ Link []domain.Link })

	queryParams := repository.NewLinkQueryParams(&page, &limit)
	links, err := ls.repo.GetMany(queryParams)
	if err != nil {
		return result, err
	}

	result.Link = links
	return result, nil
}

// CRUD get one
func (ls linkService) GetById(userId, id uint) (
	*struct{ Link domain.Link },
	error,
) {
	result := new(struct{ Link domain.Link })

	link, err := ls.repo.GetById(id)
	if err != nil {
		return result, err
	}

	result.Link = link
	return result, nil
}

// CRUD create
func (ls linkService) Create(
	userId uint,
	destination string,
) error {
	now := time.Now()
	link, err := domain.NewLink(
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

	if err := ls.repo.Create(link); err != nil {
		return err
	}

	// Launch link.published event or something

	return nil
}

// CRUD update
func (ls linkService) UpdateById(userId, id uint) error {
	return nil
}

// CRUD delete
func (ls linkService) Delete(userId, id uint) error {
	link, err := ls.repo.GetById(id)
	if err != nil {
		return err
	}
	if link.UserId() != userId {
		return oops.Unauthorized{
			Err: errors.New("You don't have access to this link")}
	}

	return ls.repo.DeleteById(id)
}

// ===================================
// Non-CRUD
// ===================================

func (ls linkService) Redirect(shortened string) error {
	return nil
}
