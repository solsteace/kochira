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

type Status struct {
	service service.Status
}

func NewStatus(service service.Status) Status {
	return Status{service: service}
}

func (s Status) FindSelf(w http.ResponseWriter, r *http.Request) error {
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

func (s Status) InitSubscription(msg amqp091.Delivery) error {
	payload, err := messaging.DeCreateSubscription(msg.Body)
	if err != nil {
		return fmt.Errorf("<consume callback>: %w", err)
	}

	if err := s.service.Init(payload.Data.Users); err != nil {
		return fmt.Errorf("<consume callback>: %w", err)
	}
	return nil
}
