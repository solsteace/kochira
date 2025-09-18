package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/subscription/internal/controller"
	"github.com/solsteace/kochira/subscription/internal/middleware"
)

type Subscription struct {
	controller  controller.Subscription
	userContext middleware.UserContext
}

func (sr Subscription) Use(parent *chi.Mux) {
	subscription := chi.NewRouter()
	subscription.Group(func(r chi.Router) {
		r.Use(sr.userContext.Handle)
		r.Get("/", reqres.HttpHandlerWithError(sr.controller.FindSelf))
	})
	parent.Mount("/subscription", subscription)
}

func NewSubscription(
	controller controller.Subscription,
	userContext middleware.UserContext,
) Subscription {
	return Subscription{controller, userContext}
}
