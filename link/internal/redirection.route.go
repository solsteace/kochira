package internal

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
)

type redirectionRoute struct {
	service linkService
}

func (rr redirectionRoute) Use(parent *chi.Mux) {
	parent.Get("/{shortened}", reqres.HttpHandlerWithError(rr.redirect))
}

func (rr redirectionRoute) redirect(w http.ResponseWriter, r *http.Request) error {
	shortened := chi.URLParam(r, "shortened")
	destination, err := rr.service.Redirect(shortened)
	if err != nil {
		return err
	}

	http.Redirect(w, r, destination, http.StatusTemporaryRedirect)
	return nil
}

func NewRedirectionRoute(service linkService) redirectionRoute {
	return redirectionRoute{service}
}
