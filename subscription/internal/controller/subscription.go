package controller

import (
	"fmt"
	"net/http"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rabbitmq/amqp091-go"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/go-lib/temporary/messaging"
	"github.com/solsteace/kochira/subscription/internal/middleware"
	"github.com/solsteace/kochira/subscription/internal/service"
)

type Subscription struct {
	service service.Subscription
}

func NewSubscription(service service.Subscription) Subscription {
	return Subscription{service: service}
}

func (s Subscription) FindSelf(w http.ResponseWriter, r *http.Request) error {
	reqId := chiMiddleware.GetReqID(r.Context())
	userId, _ := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	result, err := s.service.GetByUserId(uint64(userId))
	if err != nil {
		return fmt.Errorf("[%s] internal<subscriptionRoute.Use>: %w", reqId, err)
	}

	return reqres.HttpOk(w, http.StatusOK, map[string]any{
		"id":        result.Id(),
		"userId":    result.UserId(),
		"expiredAt": result.ExpiredAt()})
}

func (s Subscription) InitSubscription(msg amqp091.Delivery) error {
	payload, err := messaging.DeCreateSubscription(msg.Body)
	if err != nil {
		return fmt.Errorf("<consume callback>: %w", err)
	}

	if err := s.service.Init(payload.Data.Users); err != nil {
		return fmt.Errorf("<consume callback>: %w", err)
	}
	return nil
}
