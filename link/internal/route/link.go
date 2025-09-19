package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/link/internal/controller"
	"github.com/solsteace/kochira/link/internal/middleware"
)

type link struct {
	controller  controller.Link
	userContext middleware.UserContext
}

func (l link) Use(parent *chi.Mux) {
	link := chi.NewRouter()
	link.Group(func(r chi.Router) {
		r.Use(l.userContext.Handle)
		r.Get("/my", reqres.HttpHandlerWithError(l.controller.GetSelf))
		r.Get("/my/{id}", reqres.HttpHandlerWithError(l.controller.GetById))
		r.Post("/", reqres.HttpHandlerWithError(l.controller.Create))
		r.Put("/{id}", reqres.HttpHandlerWithError(l.controller.UpdateById))
		r.Delete("/{id}", reqres.HttpHandlerWithError(l.controller.DeleteById))
	})
	parent.Mount("/link", link)
}

func NewLink(controller controller.Link, userContext middleware.UserContext) link {
	return link{controller, userContext}
}
