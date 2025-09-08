package command

import "time"

type initializeLinkData struct {
	Id        uint64    `json:"id"`
	ExpiredAt time.Time `json:"expiredAt"`
}

type InitializeLink struct {
	Meta meta               `json:"meta"`
	Data initializeLinkData `json:"data"`
}

func NewInitializeLink(id uint64, expiredAt time.Time) InitializeLink {
	return InitializeLink{
		Meta: newMeta(),
		Data: initializeLinkData{
			Id:        id,
			ExpiredAt: expiredAt}}
}
