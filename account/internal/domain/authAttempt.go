package domain

import "time"

type AuthAttempt struct {
	ok   bool
	time time.Time
}

func NewAuthAttempt(ok bool, time time.Time) (AuthAttempt, error) {
	return AuthAttempt{ok, time}, nil
}

func (aa AuthAttempt) Ok() bool {
	return aa.ok
}

func (aa AuthAttempt) Time() time.Time {
	return aa.time
}
