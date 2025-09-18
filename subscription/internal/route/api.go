package route

import (
	"net/http"
	"time"

	"github.com/solsteace/go-lib/reqres"
)

type Api struct {
	upSince int64
}

func (a Api) Health(w http.ResponseWriter, r *http.Request) error {
	return reqres.HttpOk(
		w,
		http.StatusOK,
		map[string]any{
			"msg": "Server is healthy",
			"data": map[string]any{
				"uptime": time.Now().Unix() - a.upSince,
			}})
}

func (a Api) NotFound(w http.ResponseWriter, r *http.Request) error {
	return reqres.HttpOk(
		w,
		http.StatusNotFound,
		map[string]any{
			"msg": "The endpoint you're reaching wasn't found"})
}

func NewApi(upSince int64) Api {
	return Api{upSince}
}
