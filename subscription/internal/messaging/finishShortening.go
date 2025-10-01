package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/solsteace/kochira/subscription/internal/domain/subscription/messaging"
)

type finishShorteningPerkData struct {
	Limit          uint          `json:"limit"`          // How many simultaneous-active-links a user could make at a time?
	Lifetime       time.Duration `json:"lifetime"`       // How long a link would last since its shortening?
	AllowShortEdit bool          `json:"allowShortEdit"` // Does the user allowed to edit the shortened link?
}

type finishShorteningData struct {
	Id        uint64                   `json:"id"`        // What is the id of this message?
	ContextId uint64                   `json:"contextId"` // What is the id of the message's context? (determined externally)
	Usecase   string                   `json:"usecase"`   // What is the purpose of the check? (determined externally)
	Perk      finishShorteningPerkData `json:"perk"`
}

type FinishShorteningMessenger struct {
	Version uint
}

func (fsm FinishShorteningMessenger) FromSubscriptionChecked(
	msg messaging.SubscriptionChecked,
) ([]byte, error) {
	payload := struct {
		Meta meta                 `json:"meta"`
		Data finishShorteningData `json:"data"`
	}{
		Meta: meta{
			Version:  fsm.Version,
			IssuedAt: time.Now()},
		Data: finishShorteningData{
			Id:        msg.Id(),
			ContextId: msg.ContextId(),
			Usecase:   msg.Usecase(),
			Perk: finishShorteningPerkData{
				Limit:          msg.Limit(),
				Lifetime:       msg.Lifetime(),
				AllowShortEdit: msg.AllowShortEdit()},
		},
	}

	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		return []byte{}, fmt.Errorf(
			"messaging<finishShorteningMessenger.DeFinishShortening>: %w", err)
	}
	return marshalledPayload, nil
}
