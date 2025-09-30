package messaging

import (
	"encoding/json"
	"fmt"
)

type createSubscriptionData struct {
	Users []uint64 `json:"users"`
}

type createSubscriptionPayload struct {
	Meta meta                   `json:"meta"`
	Data createSubscriptionData `json:"data"`
}

type CreateSubscriptionMessenger struct {
	Version uint
}

func (csm CreateSubscriptionMessenger) FromMsg(msg []byte) (*createSubscriptionPayload, error) {
	payload := new(createSubscriptionPayload)
	if err := json.Unmarshal(msg, payload); err != nil {
		return nil, fmt.Errorf(
			"messaging<CreateSubscriptionMessenger.FromMsg>: %w", err)
	}
	return payload, nil
}
