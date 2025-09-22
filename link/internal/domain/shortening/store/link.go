package store

import "github.com/solsteace/kochira/link/internal/domain/shortening"

type Link[queryParams any] interface {
	// Queries ============

	// Retrieves many links
	GetMany(q queryParams) ([]shortening.Link, error)

	GetManyByUser(userId uint64, q queryParams) ([]shortening.Link, error)

	GetById(id uint64) (shortening.Link, error)

	// Retrieves the number of links owned by user
	CountByUserId(userId uint64) (uint, error)

	// Commands ===========

	Load(id uint64) (shortening.Link, error)
	Create(l shortening.Link) (uint64, error)
	Update(l shortening.Link) error
	DeleteById(id uint64) error
}
