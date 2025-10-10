package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	shorteningMsg "github.com/solsteace/kochira/link/internal/domain/shortening/messaging"
	"github.com/solsteace/kochira/link/internal/messaging"
	"github.com/solsteace/kochira/link/internal/middleware"
	"github.com/solsteace/kochira/link/internal/service"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type Shortening struct {
	service             service.Shortening
	checkSubscription   messaging.CheckSubscriptionMessenger
	finishShortening    messaging.FinishShorteningMessenger
	subscriptionExpired messaging.SubscriptionExpiredMessenger
}

// Move later to a viewer object or something
type shorteningLinkView struct {
	Id          uint64    `json:"id"`
	UserId      uint64    `json:"user_id"`
	Shortened   string    `json:"shortened"`
	Alias       string    `json:"alias"`
	Destination string    `json:"destination"`
	IsOpen      bool      `json:"is_open"`
	UpdatedAt   time.Time `json:"updated_at"`
	ExpiredAt   time.Time `json:"expired_at"`
}

func (lr Shortening) GetSelf(w http.ResponseWriter, r *http.Request) error {
	reqId := chiMiddleware.GetReqID(r.Context())
	var limit *uint
	var page *uint
	rq := r.URL.Query()
	qLimit, err := strconv.ParseUint(rq.Get("limit"), 10, 64)
	if err == nil {
		temp := uint(qLimit)
		limit = &temp
	}
	qPage, err := strconv.ParseUint(rq.Get("page"), 10, 64)
	if err == nil {
		temp := uint(qPage)
		page = &temp
	}

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	result, err := lr.service.GetSelf(uint64(userId), page, limit)
	if err != nil {
		return fmt.Errorf("[%s] controller<Shortening.GetSelf>: %w", reqId, err)
	}

	resPayload := []shorteningLinkView{}
	for _, r := range result {
		resPayload = append(resPayload, shorteningLinkView{
			Id:          r.Id(),
			UserId:      r.UserId(),
			Shortened:   r.Shortened(),
			Alias:       r.Alias(),
			Destination: r.Destination(),
			IsOpen:      r.IsOpen(),
			UpdatedAt:   r.UpdatedAt(),
			ExpiredAt:   r.ExpiredAt()})
	}
	if err := reqres.HttpOk(w, http.StatusOK, resPayload); err != nil {
		return fmt.Errorf("[%s] controller<Shortening.GetSelf>: %w", reqId, err)
	}
	return nil
}

func (lr Shortening) GetById(w http.ResponseWriter, r *http.Request) error {
	reqId := chiMiddleware.GetReqID(r.Context())
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return fmt.Errorf("[%s] controller<Shortening.GetById>: %w", reqId, err)
	}

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	result, err := lr.service.GetById(uint64(userId), id)
	if err != nil {
		return fmt.Errorf("[%s] controller<Shortening.GetById>: %w", reqId, err)
	}

	resPayload := shorteningLinkView{
		Id:          result.Id(),
		UserId:      result.UserId(),
		Shortened:   result.Shortened(),
		Alias:       result.Alias(),
		Destination: result.Destination(),
		IsOpen:      result.IsOpen(),
		UpdatedAt:   result.UpdatedAt(),
		ExpiredAt:   result.ExpiredAt()}
	if err := reqres.HttpOk(w, http.StatusOK, resPayload); err != nil {
		return fmt.Errorf("[%s] controller<Shortening.GetById>: %w", reqId, err)
	}
	return nil
}

func (lr Shortening) Create(w http.ResponseWriter, r *http.Request) error {
	reqId := chiMiddleware.GetReqID(r.Context())
	reqPayload := new(struct {
		Destination string `json:"destination"`
	})
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return fmt.Errorf("[%s] controller<Shortening.Create>: %w", reqId, err)
	}
	defer r.Body.Close()

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	err := lr.service.Create(uint64(userId), reqPayload.Destination)
	if err != nil {
		return fmt.Errorf("[%s] controller<Shortening.Create>: %w", reqId, err)
	}

	if err := reqres.HttpOk(w, http.StatusCreated, nil); err != nil {
		return fmt.Errorf("[%s] controller<Shortening.Create>: %w", reqId, err)
	}
	return nil
}

func (lr Shortening) UpdateById(w http.ResponseWriter, r *http.Request) error {
	reqId := chiMiddleware.GetReqID(r.Context())
	reqPayload := new(struct {
		Alias       string `json:"alias"`
		Destination string `json:"destination"`
		IsOpen      bool   `json:"isOpen"`
	})
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return fmt.Errorf("[%s] controller<Shortening.UpdateById>: %w", reqId, err)
	}
	defer r.Body.Close()

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return fmt.Errorf("[%s] controller<Shortening.UpdateById>: %w", reqId, err)
	}

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	err = lr.service.UpdateById(
		uint64(userId),
		id,
		reqPayload.Alias,
		reqPayload.Destination,
		reqPayload.IsOpen)
	if err != nil {
		return fmt.Errorf("[%s] controller<Shortening.UpdateById>: %w", reqId, err)
	}

	if err := reqres.HttpOk(w, http.StatusOK, nil); err != nil {
		return fmt.Errorf("[%s] controller<Shortening.Create>: %w", reqId, err)
	}
	return nil
}

func (lr Shortening) DeleteById(w http.ResponseWriter, r *http.Request) error {
	reqId := chiMiddleware.GetReqID(r.Context())
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return fmt.Errorf("[%s] controller<Shortening.DeleteById>: %w", reqId, err)
	}

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	if err := lr.service.DeleteById(uint64(userId), id); err != nil {
		return fmt.Errorf("[%s] controller<Shortening.DeleteById>: %w", reqId, err)
	}

	if err := reqres.HttpOk(w, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("[%s] controller<Shortening.Create>: %w", reqId, err)
	}
	return nil
}

// ===============================
// Event handling
// ===============================

func (sc Shortening) ListenFinishShortening(msg []byte) error {
	payload, err := sc.finishShortening.FromMsg(msg)
	if err != nil {
		return fmt.Errorf("controller<Shortening.ListenFinishShortening>: %w", err)
	}

	if sc.finishShortening.Version != payload.Meta.Version {
		return fmt.Errorf(
			"controller<Shortening.ListenFinishShortening>: "+
				"incompatible version between messenger(v:%d) and message(v:%d)",
			sc.finishShortening.Version, payload.Meta.Version)
	}

	switch payload.Data.Usecase {
	case shorteningMsg.LinkShortenedName:
		err = sc.service.HandleLinkShortened(
			payload.Data.ContextId,
			payload.Data.Perk.Lifetime,
			payload.Data.Perk.Limit)
		if err2 := sc.service.CompensateLinkShortened(payload.Data.ContextId); err != nil {
			err = fmt.Errorf("%w [triggered by: %w]", err2, err)
		}
	case shorteningMsg.ShortConfiguredName:
		err = sc.service.HandleShortConfigured(
			payload.Data.ContextId,
			payload.Data.Perk.AllowShortEdit)
	}
	if err != nil {
		return fmt.Errorf("controller<Shortening.ListenFinishShortening>: %w", err)
	}
	return nil
}

func (sc Shortening) ListenSubscriptionExpired(msg []byte) error {
	payload, err := sc.subscriptionExpired.FromMsg(msg)
	if err != nil {
		return fmt.Errorf("controller<Shortening.ListenSubscriptionExpired>: %w", err)
	}

	if sc.subscriptionExpired.Version != payload.Meta.Version {
		return fmt.Errorf(
			"controller<Shortening.ListenFinishShortening>: "+
				"incompatible version between messenger(v:%d) and message(v:%d)",
			sc.finishShortening.Version, payload.Meta.Version)
	}

	err = sc.service.HandleSubscriptionExpired(
		payload.Data.UserId,
		payload.Data.Perk.Limit,
		payload.Data.Perk.AllowShortEdit)
	if err != nil {
		return fmt.Errorf("controller<Shortening.ListenSubscriptionExpired>: %w", err)
	}
	return nil
}

// Creates new `Shortening` and initiates essentials for messaging purposes
func NewShortening(service service.Shortening) Shortening {
	return Shortening{
		service:             service,
		checkSubscription:   messaging.CheckSubscriptionMessenger{Version: 1},
		finishShortening:    messaging.FinishShorteningMessenger{Version: 1},
		subscriptionExpired: messaging.SubscriptionExpiredMessenger{Version: 1}}
}
