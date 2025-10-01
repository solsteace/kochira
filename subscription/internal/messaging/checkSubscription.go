package messaging

import (
	"encoding/json"
	"fmt"
)

type checkSubscriptionData struct {
	Id      uint64 `json:"id"`      // What is the id of the message?
	UserId  uint64 `json:"userId"`  // Whose subscription to check?
	Usecase string `json:"usecase"` // What is the purpose of the check?
}

type checkSubscriptionPayload struct {
	Meta meta                  `json:"meta"`
	Data checkSubscriptionData `json:"data"`
}

type CheckSubscriptionMessenger struct {
	Version uint
}

func (csm CheckSubscriptionMessenger) FromMsg(msg []byte) (*checkSubscriptionPayload, error) {
	payload := new(checkSubscriptionPayload)
	if err := json.Unmarshal(msg, payload); err != nil {
		return nil, fmt.Errorf("messaging<CheckSubscriptionMessenger.FromMsg>: %w", err)
	}
	return payload, nil
}
