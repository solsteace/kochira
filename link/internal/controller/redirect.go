package controller

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/solsteace/kochira/link/internal/service"
)

type Redirect struct {
	service service.Redirect
}

func (rc Redirect) Go(w http.ResponseWriter, r *http.Request) error {
	reqId := chiMiddleware.GetReqID(r.Context())
	shortened := chi.URLParam(r, "shortened")
	destination, err := rc.service.Go(shortened)
	if err != nil {
		return fmt.Errorf("[%s] controller<Redirection.Go>: %w", reqId, err)
	}

	http.Redirect(w, r, destination, http.StatusTemporaryRedirect)
	return nil
}

func NewRedirect(service service.Redirect) Redirect {
	return Redirect{service}
}
