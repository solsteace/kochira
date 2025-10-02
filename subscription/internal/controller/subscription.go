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
	checkSubscription  messaging.CheckSubscriptionMessenger
}

func NewSubscription(
	service service.Subscription,
	createSubscription messaging.CreateSubscriptionMessenger,
	checkSubscription messaging.CheckSubscriptionMessenger,
) Subscription {
	return Subscription{
		service:            service,
		createSubscription: createSubscription,
		checkSubscription:  checkSubscription}
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

func (s Subscription) ListenCreateSubscription(msg []byte) error {
	payload, err := s.createSubscription.FromMsg(msg)
	if err != nil {
		return fmt.Errorf("controller<Subscription.ListenCreateSubscription>: %w", err)
	}

	// TODO: un-bypass
	if s.createSubscription.Version != payload.Meta.Version && false {
		return fmt.Errorf(
			"controller<Subscription.ListenCreateSubscription>: "+
				"incompatible version between messenger(v:%d) and message(v:%d)",
			s.createSubscription.Version, payload.Meta.Version)
	}

	if err := s.service.Init(payload.Data.Users); err != nil {
		return fmt.Errorf("controller<Subscription.ListenCreateSubscription>: %w", err)
	}
	return nil
}

func (s Subscription) ListenCheckSubscription(msg []byte) error {
	payload, err := s.checkSubscription.FromMsg(msg)
	if err != nil {
		return fmt.Errorf("controller<Subscription.ListenCheckSubscription>: %w", err)
	}

	if s.checkSubscription.Version != payload.Meta.Version {
		return fmt.Errorf(
			"controller<Subscription.ListenCheckSubscription>: "+
				"incompatible version between messenger(v:%d) and message(v:%d)",
			s.createSubscription.Version, payload.Meta.Version)
	}

	err = s.service.Check(
		payload.Data.UserId,
		payload.Data.CtxId,
		payload.Data.Usecase)
	if err != nil {
		return fmt.Errorf("controller<Subscription.ListenCheckSubscription>: %w", err)
	}
	return nil
}
