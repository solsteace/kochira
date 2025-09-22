package store

import "github.com/solsteace/kochira/link/internal/domain/redirect"

type Shortening interface {
	GetByShortened(shortened string) (redirect.Link, error)
}
