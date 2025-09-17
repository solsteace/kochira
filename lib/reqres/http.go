package reqres

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/solsteace/go-lib/oops/adapter"
)

type httpHandlerWithError = func(w http.ResponseWriter, r *http.Request) error

// Kudos: https://boldlygo.tech/posts/2024-01-08-error-handling/
func HttpHandlerWithError(fx httpHandlerWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fx(w, r); err != nil {
			log.Println(err)
			HttpErr(w, err)
		}
	}
}

func httpRespondWithJSON(w http.ResponseWriter, statusCode int, payload any) error {
	resPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	w.WriteHeader(statusCode)
	w.Write(resPayload)
	return nil
}

func HttpOk(w http.ResponseWriter, statusCode int, payload any) error {
	return httpRespondWithJSON(w, statusCode, payload)
}

func HttpErr(w http.ResponseWriter, err error) error {
	statusCode := adapter.HttpStatusCode(err)
	msg := adapter.HttpErrorMsg(err)
	payload := map[string]any{"msg": msg}
	return httpRespondWithJSON(w, statusCode, payload)
}
