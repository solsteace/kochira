package internal

import (
	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
)

func UseRedirection(
	parent *chi.Mux,
	controller linkController,
) {
	parent.Get("/{shortened}", reqres.HttpHandlerWithError(controller.Redirect))
}
