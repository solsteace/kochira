package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/link/internal/controller"
)

type redirect struct {
	controller controller.Redirect
}

func (r redirect) Use(parent *chi.Mux) {
	parent.Get("/{shortened}", reqres.HttpHandlerWithError(r.controller.Go))
}

func NewRedirect(controller controller.Redirect) redirect {
	return redirect{controller}
}
