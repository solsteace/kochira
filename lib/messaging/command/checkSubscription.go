package command

type checkSubscriptionData struct {
	UserId uint64 `json:"userId"`
}

type CheckSubscription struct {
	Meta meta                  `json:"meta"`
	Data checkSubscriptionData `json:"data"`
}

func NewCheckSubscription(userId uint64) CheckSubscription {
	return CheckSubscription{
		Meta: newMeta(),
		Data: checkSubscriptionData{UserId: userId}}
}
