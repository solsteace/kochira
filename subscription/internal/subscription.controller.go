package internal

import (
	"net/http"

	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/subscription/internal/middleware"
)

type SubscriptionController struct {
	service SubscriptionService
}

func NewSubscriptionController(service SubscriptionService) SubscriptionController {
	return SubscriptionController{service}
}

func (sc SubscriptionController) FindSelf(w http.ResponseWriter, r *http.Request) error {
	userId, _ := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	result, err := sc.service.GetByUserId(uint64(userId))
	if err != nil {
		return err
	}

	return reqres.HttpOk(w, http.StatusOK, map[string]any{
		"id":        result.Id(),
		"userId":    result.UserId(),
		"expiredAt": result.ExpiredAt()})
}
