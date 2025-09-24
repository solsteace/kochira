package messaging

import (
	"encoding/json"
	"fmt"
	"time"
)

type finishShorteningData struct {
	Id     uint64 `json:"id"`
	UserId uint64 `json:"userId"`
	Perk   struct {
		Limit     uint          `json:"limit"`
		Lifetime  time.Duration `json:"lifetime"`
		AllowEdit bool          `json:"allowEdit"`
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
