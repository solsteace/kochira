package store

import "github.com/solsteace/kochira/account/internal/domain/auth"

type Attempt interface {
	Add(userId uint, row auth.Attempt) error
	Get(userId uint) ([]auth.Attempt, error)
}
