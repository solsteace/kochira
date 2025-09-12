package messaging

import "encoding/json"

type createSubscriptionData struct {
	Users []uint64 `json:"users"`
}

type createSubscription struct {
	Meta meta                   `json:"meta"`
	Data createSubscriptionData `json:"Data"`
}

func SerCreateSubscription(users []uint64) ([]byte, error) {
	msg := createSubscription{
		Meta: newMeta(),
		Data: createSubscriptionData{
			Users: users}}
	return json.Marshal(msg)
}

func DeCreateSubscription(bytes []byte) (*createSubscription, error) {
	msg := new(createSubscription)
	if err := json.Unmarshal(bytes, msg); err != nil {
		return &createSubscription{}, err
	}
	return msg, nil
}
