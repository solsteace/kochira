package view

import "time"

type Link struct {
	Id          uint64
	UserId      uint64
	Shortened   string
	Destination string
	IsOpen      bool
	UpdatedAt   time.Time
	ExpiredAt   time.Time
}
