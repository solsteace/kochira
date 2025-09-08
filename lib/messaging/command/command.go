package command

import "time"

type meta struct {
	IssuedAt time.Time `json:"issuedAt"`
}

func newMeta() meta {
	return meta{
		IssuedAt: time.Now(),
	}
}
