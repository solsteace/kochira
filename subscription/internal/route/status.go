package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/subscription/internal/controller"
	"github.com/solsteace/kochira/subscription/internal/middleware"
)

type status struct {
	controller  controller.Status
	userContext middleware.UserContext
}

func (s status) Use(parent *chi.Mux) {
	status := chi.NewRouter()
	status.Group(func(r chi.Router) {
		r.Use(s.userContext.Handle)
		r.Get("/", reqres.HttpHandlerWithError(s.controller.FindSelf))
	})
	parent.Mount("/status", status)
}

func NewStatus(
	controller controller.Status,
	userContext middleware.UserContext,
) status {
	return status{controller, userContext}
}
