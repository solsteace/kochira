package internal

import (
	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/subscription/internal/middleware"
)

func UseSubscription(
	parent *chi.Mux,
	controller SubscriptionController,
	userContext middleware.UserContext,
) {
	subscription := chi.NewRouter()
	subscription.Group(func(r chi.Router) {
		r.Use(userContext.Handle)
		r.Get("/", reqres.HttpHandlerWithError(controller.FindSelf))
	})
	parent.Mount("/subscription", subscription)
}
