package messaging

import (
	"encoding/json"
	"fmt"
)

type subscriptionExpiredPerkData struct {
	Limit          uint `json:"limit"`          // How many simultaneous-active-links a user could make at a time?
	AllowShortEdit bool `json:"allowShortEdit"` // Does the user allowed to edit the shortened link?
}

type subscriptionExpiredData struct {
	Id     uint64                      `json:"id"`     // What is the id of this message?
	UserId uint64                      `json:"userId"` // Whose subscription had expired?
	Perk   subscriptionExpiredPerkData `json:"perk"`
}

type subscriptionExpiredPayload struct {
	Meta meta                    `json:"meta"`
	Data subscriptionExpiredData `json:"data"`
}

type SubscriptionExpiredMessenger struct {
	Version uint
}

func (fsm SubscriptionExpiredMessenger) FromMsg(msg []byte) (*subscriptionExpiredPayload, error) {
	payload := new(subscriptionExpiredPayload)
	if err := json.Unmarshal(msg, &payload); err != nil {
		return nil, fmt.Errorf(
			"messaging<finishShorteningMessenger.FromMsg>: %w", err)
	}
	return payload, nil
}
