package repository

import (
	"github.com/solsteace/kochira/link/internal/domain"
	"github.com/solsteace/kochira/link/internal/view"
)

type Link interface {
	// Queries ============

	// Retrieves many links
	GetMany(q linkQueryParams) ([]view.Link, error)

	GetManyByUser(userId uint64, q linkQueryParams) ([]view.Link, error)

	GetById(id uint64) (view.Link, error)

	// Retrieves the number of links owned by user
	CountByUserId(userId uint64) (uint, error)

	// Retrieves active redirection link. That is, link that is open and not expired yet
	GetByShortened(shortened string) (view.Link, error)

	// Commands ===========

	Load(id uint64) (domain.Link, error)
	Create(l domain.Link) (uint64, error)
	Update(l domain.Link) error
	DeleteById(id uint64) error
}

type linkQueryParams struct {
	page  uint
	limit uint
}

func (param linkQueryParams) Offset() uint {
	if param.page < 1 {
		return 0
	}
	return (param.page - 1) * param.limit
}

func NewLinkQueryParams(page, limit *uint) linkQueryParams {
	var actualPage uint = 1
	if page != nil && *page > 0 {
		actualPage = *page
	}

	var actualLimit uint = 10 // DEFAULT
	if limit != nil && *limit > 0 {
		actualLimit = *limit
	}

	return linkQueryParams{
		actualPage,
		actualLimit}
}
