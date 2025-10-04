package redirect

import (
	"fmt"
	"time"

	"github.com/solsteace/go-lib/oops"
)

type Link struct {
	Id          uint64
	Shortened   string
	Destination string
	IsOpen      bool
	ExpiredAt   time.Time
}

func (l Link) Access() (string, error) {
	switch {
	case !l.IsOpen:
		return "", fmt.Errorf(
			"service<Redirect.Go>: %w",
			oops.Forbidden{Msg: "This link is not opened by the owner"})
	case time.Now().Sub(l.ExpiredAt) > 0:
		return "", fmt.Errorf(
			"service<Redirect.Go>: %w",
			oops.Forbidden{Msg: "This link had already expired"})
	}
	return l.Destination, nil
}
