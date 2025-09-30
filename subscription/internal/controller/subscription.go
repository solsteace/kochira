package controller

import (
	"fmt"
	"net/http"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/subscription/internal/messaging"
	"github.com/solsteace/kochira/subscription/internal/middleware"
	"github.com/solsteace/kochira/subscription/internal/service"
)

type Subscription struct {
	service            service.Subscription
	createSubscription messaging.CreateSubscriptionMessenger
}

func NewSubscription(service service.Subscription, version uint) Subscription {
	return Subscription{
		service:            service,
		createSubscription: messaging.CreateSubscriptionMessenger{Version: version}}
}

func (s Subscription) FindSelf(w http.ResponseWriter, r *http.Request) error {
	reqId := chiMiddleware.GetReqID(r.Context())
	userId, _ := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	result, err := s.service.GetByUserId(uint64(userId))
	if err != nil {
		return fmt.Errorf("[%s] controller<Status.FindSelf>: %w", reqId, err)
	}

	return reqres.HttpOk(w, http.StatusOK, map[string]any{
		"id":        result.Id(),
		"userId":    result.UserId(),
		"expiredAt": result.ExpiredAt()})
}

func (s Subscription) InitSubscription(msg []byte) error {
	payload, err := s.createSubscription.FromMsg(msg)
	if err != nil {
		return fmt.Errorf("controller<Status.InitSubscription>: %w", err)
	}

	// TODO: un-bypass
	if s.createSubscription.Version != payload.Meta.Version && false {
		return fmt.Errorf(
			"controller<Subscription.InitSubscription>: "+
				"incompatible version between messenger(v:%d) and message(v:%d)",
			s.createSubscription.Version, payload.Meta.Version)
	}

	if err := s.service.Init(payload.Data.Users); err != nil {
		return fmt.Errorf("<consume callback>: %w", err)
	}
	return nil
}
