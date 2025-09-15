package internal

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/subscription/internal/middleware"
)

type subscriptionRoute struct {
	service     SubscriptionService
	userContext middleware.UserContext
}

func (sr subscriptionRoute) Use(parent *chi.Mux) {
	subscription := chi.NewRouter()
	subscription.Group(func(r chi.Router) {
		r.Use(sr.userContext.Handle)
		r.Get("/", reqres.HttpHandlerWithError(sr.findSelf))
	})
	parent.Mount("/subscription", subscription)
}

func (sr subscriptionRoute) findSelf(w http.ResponseWriter, r *http.Request) error {
	userId, _ := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	result, err := sr.service.GetByUserId(uint64(userId))
	if err != nil {
		return err
	}

	return reqres.HttpOk(w, http.StatusOK, map[string]any{
		"id":        result.Id(),
		"userId":    result.UserId(),
		"expiredAt": result.ExpiredAt()})
}

func NewSubscriptionRoute(
	service SubscriptionService,
	userContext middleware.UserContext,
) subscriptionRoute {
	return subscriptionRoute{service, userContext}
}
