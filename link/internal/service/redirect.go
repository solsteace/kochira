package service

import (
	"fmt"

	"github.com/solsteace/kochira/link/internal/domain/redirect/store"
)

type Redirect struct {
	store store.Shortening
}

// Redirects the user to the destination based on given shortened URI
func (ls Redirect) Go(shortened string) (string, error) {
	link, err := ls.store.GetByAlias(shortened)
	if err != nil {
		return "", fmt.Errorf("service<Redirect.Go>: %w", err)
	}

	destination, err := link.Access()
	if err != nil {
		return "", fmt.Errorf("service<Redirect.Go>: %w", err)
	}
	return destination, nil
}

func NewRedirect(store store.Shortening) Redirect {
	return Redirect{store}
}
