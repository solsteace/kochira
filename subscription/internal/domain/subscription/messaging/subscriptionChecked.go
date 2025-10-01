package messaging

import "time"

const SubscriptionCheckedName = "subscription.checked"

type SubscriptionChecked struct {
	id        uint64
	contextId uint64
	usecase   string
	lifetime  time.Duration
	limit     uint
	allowEdit bool
}

func (sc SubscriptionChecked) Id() uint64              { return sc.id }
func (sc SubscriptionChecked) ContextId() uint64       { return sc.contextId }
func (sc SubscriptionChecked) Usecase() string         { return sc.usecase }
func (sc SubscriptionChecked) Lifetime() time.Duration { return sc.lifetime }
func (sc SubscriptionChecked) Limit() uint             { return sc.limit }
func (sc SubscriptionChecked) AllowShortEdit() bool    { return sc.allowEdit }

func NewSubscriptionChecked(
	id uint64,
	contextId uint64,
	usecase string,
	lifetime time.Duration,
	limit uint,
	allowShortEdit bool,
) SubscriptionChecked {
	return SubscriptionChecked{
		id:        id,
		contextId: contextId,
		usecase:   usecase,
		lifetime:  lifetime,
		limit:     limit,
		allowEdit: allowShortEdit}
}
