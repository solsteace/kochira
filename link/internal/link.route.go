package internal

import (
	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/link/internal/middleware"
)

func UseLink(
	parent *chi.Mux,
	controller linkController,
	userContext middleware.UserContext,
) {
	link := chi.NewRouter()

	link.Group(func(r chi.Router) {
		r.Use(userContext.Handle)

		// Put user context middleware
		r.Get("/my/{id}", reqres.HttpHandlerWithError(controller.GetById))
		r.Get("/my", reqres.HttpHandlerWithError(controller.GetSelf))
		r.Post("/{id}", reqres.HttpHandlerWithError(controller.Create))
		r.Put("/{id}", reqres.HttpHandlerWithError(controller.UpdateById))
		r.Delete("/{id}", reqres.HttpHandlerWithError(controller.DeleteById))
	})
	link.Get("/{shotened}", reqres.HttpHandlerWithError(controller.Redirect))
}
