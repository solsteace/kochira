package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/link/internal/controller"
	"github.com/solsteace/kochira/link/internal/middleware"
)

type shortening struct {
	controller  controller.Shortening
	userContext middleware.UserContext
}

func (s shortening) Use(parent *chi.Mux) {
	shortening := chi.NewRouter()
	shortening.Group(func(r chi.Router) {
		r.Use(s.userContext.Handle)
		r.Get("/my", reqres.HttpHandlerWithError(s.controller.GetSelf))
		r.Get("/my/{id}", reqres.HttpHandlerWithError(s.controller.GetById))
		r.Post("/", reqres.HttpHandlerWithError(s.controller.Create))
		r.Put("/{id}", reqres.HttpHandlerWithError(s.controller.UpdateById))
		r.Delete("/{id}", reqres.HttpHandlerWithError(s.controller.DeleteById))
	})
	parent.Mount("/link", shortening)
}

func NewShortening(controller controller.Shortening, userContext middleware.UserContext) shortening {
	return shortening{controller, userContext}
}
