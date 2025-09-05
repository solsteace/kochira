package repository

import "github.com/solsteace/kochira/link/internal/domain"

type Link interface {
	GetMany(q linkQueryParams) ([]domain.Link, error)
	GetById(id uint) (domain.Link, error)
	Create(l domain.Link) error
	Update(l domain.Link) error
	DeleteById(id uint) error
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
