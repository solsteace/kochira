package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/account/internal/controller"
)

type account struct {
	controller controller.Account
}

func (ar account) Use(parent *chi.Mux) {
	auth := chi.NewRouter()
	auth.Post("/register", reqres.HttpHandlerWithError(ar.controller.Register))

	parent.Mount("/account", auth)
}

func NewAccount(controller controller.Account) account {
	return account{controller: controller}
}
