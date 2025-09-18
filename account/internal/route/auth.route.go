package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/account/internal/controller"
)

type authRoute struct {
	controller controller.Auth
}

func (ar authRoute) Use(parent *chi.Mux) {
	auth := chi.NewRouter()
	auth.Get("/infer", reqres.HttpHandlerWithError(ar.controller.Infer))
	auth.Post("/register", reqres.HttpHandlerWithError(ar.controller.Register))
	auth.Post("/login", reqres.HttpHandlerWithError(ar.controller.Login))
	auth.Post("/refresh", reqres.HttpHandlerWithError(ar.controller.Refresh))
	auth.Post("/logout", reqres.HttpHandlerWithError(ar.controller.Logout))

	parent.Mount("/auth", auth)
}

func NewAuth(controller controller.Auth) authRoute {
	return authRoute{controller}
}
