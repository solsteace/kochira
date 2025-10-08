package shortening

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"net/url"
	"time"

	"github.com/solsteace/go-lib/oops"
)

const (
	sHORTENED_MAX_LEN   = 15
	sHORTENED_CHARSET   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	dESTINATION_MAX_LEN = 255
)

type Link struct {
	id          uint64
	userId      uint64
	shortened   string
	alias       string
	destination string
	isOpen      bool
	updatedAt   time.Time
	expiredAt   time.Time
}

// Sets shortened link
func (l *Link) Shorten() {
	shortened := make([]byte, 15)
	for i, _ := range shortened {
		idx := rand.Int() % len(sHORTENED_CHARSET)
		shortened[i] = sHORTENED_CHARSET[idx]
	}
	l.shortened = string(shortened)
	l.alias = l.shortened
}
func (l *Link) Activate() {
	l.isOpen = true
}
func (l *Link) Deactivate() {
	l.isOpen = false
}

func (l Link) HadExpired() bool {
	return time.Now().After(l.expiredAt)
}
func (l Link) AccessibleBy(userId uint64) bool {
	return l.userId == userId
}
func (l Link) HasCustomAlias() bool {
	return l.shortened != l.alias
}

func (l Link) Id() uint64           { return l.id }
func (l Link) UserId() uint64       { return l.userId }
func (l Link) Shortened() string    { return l.shortened }
func (l Link) Alias() string        { return l.alias }
func (l Link) Destination() string  { return l.destination }
func (l Link) IsOpen() bool         { return l.isOpen }
func (l Link) UpdatedAt() time.Time { return l.updatedAt }
func (l Link) ExpiredAt() time.Time { return l.expiredAt }

func NewLink(
	id *uint64,
	userId uint64,
	shortened string,
	alias string,
	destination string,
	isOpen bool,
	updatedAt time.Time,
	expiredAt time.Time,
) (Link, error) {
	var actualId uint64 = 0
	if id != nil {
		actualId = *id
	}

	if len(shortened) > sHORTENED_MAX_LEN {
		err := oops.BadValues{
			Err: errors.New(fmt.Sprintf(
				"Shortened could only be %d chars long at maximum",
				sHORTENED_MAX_LEN))}
		return Link{}, fmt.Errorf("domain<NewLink>: %w", err)
	} else if len(destination) > dESTINATION_MAX_LEN {
		err := oops.BadValues{
			Err: errors.New(fmt.Sprintf(
				"Destination could only be %d chars long at maximum",
				dESTINATION_MAX_LEN))}
		return Link{}, fmt.Errorf("domain<NewLink>: %w", err)
	}

	destinationUrl, err := url.Parse(destination)
	if err != nil {
		return Link{}, fmt.Errorf("domain<NewLink>: %w", err)
	} else if destinationUrl.Scheme == "" {
		err := oops.BadValues{
			Err: errors.New("destination should contain URL scheme")}
		return Link{}, fmt.Errorf("domain<NewLink>: %w", err)
	}

	l := Link{
		id:          actualId,
		userId:      userId,
		shortened:   shortened,
		alias:       alias,
		destination: destination,
		isOpen:      isOpen,
		updatedAt:   updatedAt,
		expiredAt:   expiredAt}
	return l, nil
}
