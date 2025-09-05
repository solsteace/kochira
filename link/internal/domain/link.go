package domain

import "time"

type Link struct {
	id          uint
	userId      uint
	shortened   string
	destination string
	isOpen      bool
	updatedAt   time.Time
	expiredAt   time.Time
}

func (l Link) Id() uint             { return l.id }
func (l Link) UserId() uint         { return l.userId }
func (l Link) Shortened() string    { return l.shortened }
func (l Link) Destination() string  { return l.destination }
func (l Link) IsOpen() bool         { return l.isOpen }
func (l Link) UpdatedAt() time.Time { return l.updatedAt }
func (l Link) ExpiredAt() time.Time { return l.expiredAt }

func NewLink(
	id *uint,
	userId uint,
	shortened string,
	destination string,
	isOpen bool,
	updatedAt time.Time,
	expiredAt time.Time,
) (Link, error) {
	var actualId uint = 0
	if id != nil {
		actualId = *id
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
