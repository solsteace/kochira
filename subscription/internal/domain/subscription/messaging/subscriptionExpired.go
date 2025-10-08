package messaging

const SubscriptionExpiredName = "subscription.expired"

type SubscriptionExpired struct {
	id     uint64
	userId uint64
}

func (se SubscriptionExpired) Id() uint64     { return se.id }
func (se SubscriptionExpired) UserId() uint64 { return se.userId }

func NewSubscriptionExpired(
	id uint64,
	userId uint64,
) SubscriptionExpired {
	return SubscriptionExpired{
		id:     id,
		userId: userId}
}
