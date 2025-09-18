package route

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
)

type api struct {
	upSince int64
}

func (a api) Use(parent *chi.Mux) {
	parent.Get("/health", reqres.HttpHandlerWithError(
		func(w http.ResponseWriter, r *http.Request) error {
			return reqres.HttpOk(
				w,
				http.StatusOK,
				map[string]any{
					"msg": "Server is healthy",
					"data": map[string]any{
						"uptime": time.Now().Unix() - a.upSince,
					}})
		}))
	parent.NotFound(reqres.HttpHandlerWithError(
		func(w http.ResponseWriter, r *http.Request) error {
			return reqres.HttpOk(
				w,
				http.StatusNotFound,
				map[string]any{
					"msg": "The endpoint you're reaching wasn't found"})
		}))
}

func NewApi(upSince int64) api {
	return api{upSince}
}
