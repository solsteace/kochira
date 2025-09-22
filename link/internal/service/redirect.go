package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/kochira/link/internal/domain/redirect/store"
)

type Redirect struct {
	store store.Shortening
}

// Redirects the user to the destination based on given shortened URI
func (ls Redirect) Go(shortened string) (string, error) {
	link, err := ls.store.GetByShortened(shortened)
	if err != nil {
		return "", fmt.Errorf("service<Redirect.Go>: %w", err)
	}
	switch {
	case !link.IsOpen:
		err := oops.Forbidden{
			Err: errors.New("This link is not opened by the owner"),
			Msg: "This link is not opened by the owner"}
		return "", fmt.Errorf("service<Redirect.Go>: %w", err)
	case time.Now().Sub(link.ExpiredAt) > 0:
		err := oops.Forbidden{
			Err: errors.New("This link had already expired"),
			Msg: "This link had already expired"}
		return "", fmt.Errorf("service<Redirect.Go>: %w", err)
	}
	return link.Destination, nil
}

func NewRedirect(store store.Shortening) Redirect {
	return Redirect{store}
}
