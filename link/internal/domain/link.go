package domain

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"net/url"
	"time"

	"github.com/solsteace/go-lib/oops"
)

const (
	shortened_max_len   = 15
	shortened_chars     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	destination_max_len = 255
)

type Link struct {
	id          uint64
	userId      uint64
	shortened   string
	destination string
	isOpen      bool
	updatedAt   time.Time
	expiredAt   time.Time
}

// Sets shortened link
func (l *Link) NewShortened() {
	shortened := make([]byte, 15)
	for i, _ := range shortened {
		idx := rand.Int() % len(shortened_chars)
		shortened[i] = shortened_chars[idx]
	}
	l.shortened = string(shortened)
}

func (l Link) HadExpired() bool {
	return time.Now().After(l.expiredAt)
}

func (l Link) Id() uint64           { return l.id }
func (l Link) UserId() uint64       { return l.userId }
func (l Link) Shortened() string    { return l.shortened }
func (l Link) Destination() string  { return l.destination }
func (l Link) IsOpen() bool         { return l.isOpen }
func (l Link) UpdatedAt() time.Time { return l.updatedAt }
func (l Link) ExpiredAt() time.Time { return l.expiredAt }

func NewLink(
	id *uint64,
	userId uint64,
	shortened string,
	destination string,
	isOpen bool,
	updatedAt time.Time,
	expiredAt time.Time,
) (Link, error) {
	var actualId uint64 = 0
	if id != nil {
		actualId = *id
	}

	switch {
	case len(shortened) > shortened_max_len:
		err := oops.BadValues{
			Err: errors.New(fmt.Sprintf(
				"Shortened could only be %d chars long at maximum",
				shortened_max_len))}
		return Link{}, fmt.Errorf("domain<NewLink>: %w", err)
	case len(destination) > destination_max_len:
		err := oops.BadValues{
			Err: errors.New(fmt.Sprintf(
				"Destination could only be %d chars long at maximum",
				destination_max_len))}
		return Link{}, fmt.Errorf("domain<NewLink>: %w", err)
	}

	destinationUrl, err := url.Parse(destination)
	if err != nil {
		return Link{}, fmt.Errorf("domain<NewLink>: %w", err)
	}
	if destinationUrl.Scheme == "" {
		err := oops.BadValues{
			Err: errors.New("destination should contain URL scheme")}
		return Link{}, fmt.Errorf("domain<NewLink>: %w", err)
	}

	l := Link{
		actualId,
		userId,
		shortened,
		destination,
		isOpen,
		updatedAt,
		expiredAt}
	return l, nil
}
