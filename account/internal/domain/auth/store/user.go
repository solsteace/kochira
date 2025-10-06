package store

import "github.com/solsteace/kochira/account/internal/domain/auth"

type User interface {
	GetByUsername(username string) (auth.User, error)
}
