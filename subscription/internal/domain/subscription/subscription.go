package subscription

import "time"

type Subscription struct {
	id        uint64
	userId    uint64
	expiredAt time.Time
}

func (s Subscription) Id() uint64           { return s.id }
func (s Subscription) UserId() uint64       { return s.userId }
func (s Subscription) ExpiredAt() time.Time { return s.expiredAt }

func (s *Subscription) Extend(d time.Duration) {
	s.expiredAt = s.expiredAt.Add(d)
}

func NewSubscription(
	id *uint64,
	userId uint64,
	expiredAt time.Time,
) (Subscription, error) {
	var actualId uint64 = 0
	if id != nil {
		actualId = *id
	}

	s := Subscription{
		id:        actualId,
		userId:    userId,
		expiredAt: expiredAt}
	return s, nil
}
