package store

import "github.com/solsteace/kochira/link/internal/domain/redirect"

type Shortening interface {
	GetByAlias(shortened string) (redirect.Link, error)
}
