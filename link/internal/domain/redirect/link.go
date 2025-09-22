package redirect

import "time"

type Link struct {
	Id          uint64
	Shortened   string
	Destination string
	IsOpen      bool
	ExpiredAt   time.Time
}

func (l Link) IsRedirectable() bool {
	diff := time.Now().Sub(l.ExpiredAt)
	return l.IsOpen && diff < 0
}
