package messaging

import (
	"time"
)

type meta struct {
	Version  uint      `json:"version"`          // What is the version of this message?
	IssuedAt time.Time `json:"issuedAt"`         // When was this message issued?
	Source   string    `json:"source,omitempty"` // What is the original message this message was derived from, if any?
}
