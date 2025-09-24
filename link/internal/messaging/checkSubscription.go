package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	shorteningMsg "github.com/solsteace/kochira/link/internal/domain/shortening/messaging"
)

type checkSubscriptionData struct {
	Id     uint64 `json:"id"`
	UserId uint64 `json:"userId"`
}

// Handles integration event for commanding subscription check
type CheckSubscriptionMessenger struct {
	Version uint
}

// Transforms `linkShortened` event
func (csm CheckSubscriptionMessenger) FromLinkShortened(
	msg shorteningMsg.LinkShortened,
) ([]byte, error) {
	payload := struct {
		Meta meta                  `json:"meta"`
		Data checkSubscriptionData `json:"data"`
	}{
		Meta: meta{
			Version:  csm.Version,
			Source:   shorteningMsg.LinkShortenedName,
			IssuedAt: time.Now()},
		Data: checkSubscriptionData{
			Id:     msg.Id(),
			UserId: msg.UserId()}}

	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		return []byte{}, fmt.Errorf(
			"messaging<CheckSubscriptionMessenger.SerLinkShortened>: %w", err)
	}
	return marshalledPayload, nil
}

// Transforms `shortConfigured` event
func (csm CheckSubscriptionMessenger) FromShortConfigured(
	msg shorteningMsg.ShortConfigured,
) ([]byte, error) {
	payload := struct {
		Meta meta                  `json:"meta"`
		Data checkSubscriptionData `json:"data"`
	}{
		Meta: meta{
			Version:  csm.Version,
			Source:   shorteningMsg.ShortConfiguredName,
			IssuedAt: time.Now()},
		Data: checkSubscriptionData{
			Id:     msg.Id(),
			UserId: msg.UserId()}}

	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		return []byte{}, fmt.Errorf(
			"messaging<CheckSubscriptionMessenger.SerShortConfigured>: %w", err)
	}
	return marshalledPayload, nil
}
