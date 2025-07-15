package account

import (
	"github.com/go-chi/chi/v5"
)

func newAuthRouter(controller AuthController) chi.Router {
	r := chi.NewRouter()
	r.Post("/login", controller.login)
	r.Post("/register", controller.register)

	return r
}
