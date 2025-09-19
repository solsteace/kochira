package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/solsteace/kochira/link/internal/service"
)

type Redirection struct {
	service service.Link
}

func (re Redirection) Redirect(w http.ResponseWriter, r *http.Request) error {
	shortened := chi.URLParam(r, "shortened")
	destination, err := re.service.Redirect(shortened)
	if err != nil {
		return err
	}

	http.Redirect(w, r, destination, http.StatusTemporaryRedirect)
	return nil
}

func NewRedirection(service service.Link) Redirection {
	return Redirection{service}
}
