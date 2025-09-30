package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/subscription/internal/controller"
	"github.com/solsteace/kochira/subscription/internal/middleware"
)

type subscription struct {
	controller  controller.Subscription
	userContext middleware.UserContext
}

func (s subscription) Use(parent *chi.Mux) {
	status := chi.NewRouter()
	status.Group(func(r chi.Router) {
		r.Use(s.userContext.Handle)
		r.Get("/", reqres.HttpHandlerWithError(s.controller.FindSelf))
	})
	parent.Mount("/status", status)
}

func NewSubscription(
	controller controller.Subscription,
	userContext middleware.UserContext,
) subscription {
	return subscription{controller, userContext}
}
