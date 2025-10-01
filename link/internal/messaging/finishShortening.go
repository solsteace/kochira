package messaging

import (
	"encoding/json"
	"fmt"
	"time"
)

type finishShorteningData struct {
	Id        uint64 `json:"id"`        // What is the id of the message?
	ContextId uint64 `json:"contextId"` // What is the id of the message's context?
	Usecase   string `json:"usecase"`   // What is the purpose of the check?
	Perk      struct {
		Limit          uint          `json:"limit"`          // How many simultaneous-active-links a user could make at a time?
		Lifetime       time.Duration `json:"lifetime"`       // How long a link would last since its shortening?
		AllowShortEdit bool          `json:"allowShortEdit"` // Does the user allowed to edit the shortened link?
	} `json:"perk"`
}

type finishShorteningPayload struct {
	Meta meta                 `json:"meta"`
	Data finishShorteningData `json:"data"`
}

type FinishShorteningMessenger struct {
	Version uint
}

func (fsm FinishShorteningMessenger) FromMsg(msg []byte) (*finishShorteningPayload, error) {
	payload := new(finishShorteningPayload)
	if err := json.Unmarshal(msg, &payload); err != nil {
		return nil, fmt.Errorf(
			"messaging<finishShorteningMessenger.DeFinishShortening>: %w", err)
	}
	return payload, nil
}
