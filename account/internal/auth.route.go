package internal

import (
	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
)

func UseAuth(parent *chi.Mux, controller authController) {
	auth := chi.NewRouter()
	auth.Post("/register", reqres.HttpHandlerWithError(controller.Register))
	auth.Post("/login", reqres.HttpHandlerWithError(controller.Login))
	auth.Post("/refresh", reqres.HttpHandlerWithError(controller.Refresh))
	auth.Post("/logout", reqres.HttpHandlerWithError(controller.Logout))

	parent.Mount("/auth", auth)
}
