package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/solsteace/kochira/subscription/internal/domain/subscription/messaging"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/value"
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

type SubscriptionExpiredMessenger struct {
	Version uint
}

func (fsm SubscriptionExpiredMessenger) FromSubscriptionExpired(
	msg messaging.SubscriptionExpired,
	perk value.Perk,
) ([]byte, error) {
	payload := struct {
		Meta meta                    `json:"meta"`
		Data subscriptionExpiredData `json:"data"`
	}{
		Meta: meta{
			Version:  fsm.Version,
			IssuedAt: time.Now()},
		Data: subscriptionExpiredData{
			Id:     msg.Id(),
			UserId: msg.UserId(),
			Perk: subscriptionExpiredPerkData{
				Limit:          perk.Limit(),
				AllowShortEdit: perk.AllowShortEdit()}},
	}

	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		return []byte{}, fmt.Errorf(
			"messaging<finishShorteningMessenger.DeFinishShortening>: %w", err)
	}
	return marshalledPayload, nil
}
