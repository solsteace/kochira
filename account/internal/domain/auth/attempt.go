package auth

import "time"

type Attempt struct {
	ok   bool
	time time.Time
}

func NewAttempt(ok bool, time time.Time) (Attempt, error) {
	return Attempt{ok, time}, nil
}

func (a Attempt) Ok() bool        { return a.ok }
func (a Attempt) Time() time.Time { return a.time }
