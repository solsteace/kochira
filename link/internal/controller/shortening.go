package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

	if err := reqres.HttpOk(w, http.StatusOK, result); err != nil {
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

	if err := reqres.HttpOk(w, http.StatusOK, result); err != nil {
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
