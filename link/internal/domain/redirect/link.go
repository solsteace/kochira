package redirect

import (
	"errors"
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
		err := oops.Forbidden{
			Err: errors.New("This link is not opened by the owner"),
			Msg: "This link is not opened by the owner"}
		return "", fmt.Errorf("service<Redirect.Go>: %w", err)
	case time.Now().Sub(l.ExpiredAt) > 0:
		err := oops.Forbidden{
			Err: errors.New("This link had already expired"),
			Msg: "This link had already expired"}
		return "", fmt.Errorf("service<Redirect.Go>: %w", err)
	}
	return l.Shortened, nil

}
