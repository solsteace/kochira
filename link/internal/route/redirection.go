package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/link/internal/controller"
)

type redirection struct {
	controller controller.Link
}

func (r redirection) Use(parent *chi.Mux) {
	parent.Get("/{shortened}", reqres.HttpHandlerWithError(r.controller.Redirect))
}

func NewRedirection(controller controller.Link) redirection {
	return redirection{controller}
}
