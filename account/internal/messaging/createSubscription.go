package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/solsteace/kochira/account/internal/domain/account/messaging"
)

type createSubscriptionData struct {
	Users []uint64 `json:"users"`
}

type createSubscription struct {
	Meta meta                   `json:"meta"`
	Data createSubscriptionData `json:"Data"`
}

type CreateSubscriptionMessenger struct {
	Version uint
}

func (csm CreateSubscriptionMessenger) FromManyUserRegistered(
	msg []messaging.UserRegistered,
) ([]byte, error) {
	users := []uint64{}
	for _, u := range msg {
		users = append(users, u.UserId())
	}
	payload := struct {
		Meta meta                   `json:"meta"`
		Data createSubscriptionData `json:"Data"`
	}{
		Meta: meta{
			Version:  csm.Version,
			IssuedAt: time.Now(),
		},
		Data: createSubscriptionData{
			Users: users},
	}

	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		return []byte{}, fmt.Errorf(
			"messaging<CreateSubscriptionMessenger.FromManyUserRegistered>: %w", err)
	}
	return marshalledPayload, nil
}
