package store

import (
	"github.com/solsteace/kochira/link/internal/domain/shortening"
	"github.com/solsteace/kochira/link/internal/domain/shortening/messaging"
)

type Link[queryParams any] interface {
	// Queries ============

	GetMany(q queryParams) ([]shortening.Link, error) // Retrieves many links
	GetManyByUser(userId uint64, q queryParams) ([]shortening.Link, error)
	GetById(id uint64) (shortening.Link, error)
	CountByUserIdExcept(userId uint64, linkId uint64) (shortening.Stats, error) // Retrieves the number of links owned by user, excluding certain link

	// Commands ===========

	Create(l shortening.Link) error    // Creates Link and emits `linkShortened` message
	Configure(l shortening.Link) error // Emits `shortConfigured` message
	Update(l shortening.Link) error    // Updates link
	DeleteById(id uint64) error        // Deletes link

	// Events ===========

	GetLinkShortened(limit uint) ([]messaging.LinkShortened, error) // Retrieves pending `linkShortened` messages
	GetLinkShortenedById(id uint64) (messaging.LinkShortened, error)
	ResolveLinkShortened(id []uint64) error // Resolves pending `linkShortened` messages

	GetShortConfigured(limit uint) ([]messaging.ShortConfigured, error) // Retrieves pending `shortConfigured` messages
	GetShortConfiguredById(id uint64) (messaging.ShortConfigured, error)
	ResolveShortConfigured(id []uint64) error // Resolves pending `shortConfigured` messages
}
