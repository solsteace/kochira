package messaging

import (
	"time"
)

type meta struct {
	Version  uint      `json:"version"`  // What is the version of this message?
	IssuedAt time.Time `json:"issuedAt"` // When was this message issued?
}
