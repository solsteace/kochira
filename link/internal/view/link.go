package view

import "time"

type Link struct {
	Id          uint
	UserId      uint
	Shortened   string
	Destination string
	IsOpen      bool
	UpdatedAt   time.Time
	ExpiredAt   time.Time
}
