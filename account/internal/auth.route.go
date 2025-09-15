package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/go-lib/reqres"
)

type authRoute struct {
	service authService
}

func (ar authRoute) Use(parent *chi.Mux) {
	auth := chi.NewRouter()
	auth.Get("/infer", reqres.HttpHandlerWithError(ar.infer))
	auth.Post("/register", reqres.HttpHandlerWithError(ar.register))
	auth.Post("/login", reqres.HttpHandlerWithError(ar.login))
	auth.Post("/refresh", reqres.HttpHandlerWithError(ar.refresh))
	auth.Post("/logout", reqres.HttpHandlerWithError(ar.logout))

	parent.Mount("/auth", auth)
}

func (ac authRoute) login(w http.ResponseWriter, r *http.Request) error {
	reqPayload := new(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	})

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return err
	}
	defer r.Body.Close()

	accessToken, refreshToken, err := ac.service.Login(reqPayload.Username, reqPayload.Password)
	if err != nil {
		return err
	}

	resPayload := map[string]any{
		"token": map[string]any{
			"access":  accessToken,
			"refresh": refreshToken}}
	return reqres.HttpOk(w, http.StatusOK, resPayload)
}

func (ac authRoute) register(w http.ResponseWriter, r *http.Request) error {
	reqPayload := new(struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	})

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return err
	}
	defer r.Body.Close()

	err := ac.service.Register(
		reqPayload.Username,
		reqPayload.Password,
		reqPayload.Email)
	if err != nil {
		return err
	}

	resPayload := map[string]any{"msg": "Account successfully created"}
	return reqres.HttpOk(w, http.StatusCreated, resPayload)
}

func (ac authRoute) refresh(w http.ResponseWriter, r *http.Request) error {
	reqPayload := new(struct {
		Token string `json:"token"`
	})

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return err
	}
	defer r.Body.Close()

	accessToken, refreshToken, err := ac.service.Refresh(reqPayload.Token)
	if err != nil {
		return err
	}

	resPayload := map[string]any{
		"token": map[string]any{
			"access":  accessToken,
			"refresh": refreshToken}}
	return reqres.HttpOk(w, http.StatusOK, resPayload)
}

func (ac authRoute) logout(w http.ResponseWriter, r *http.Request) error {
	token := r.Header.Get("Authorization")
	if token == "" {
		return oops.Unauthorized{
			Err: errors.New("Auth token not found"),
			Msg: "Auth token not found"}
	}

	if err := ac.service.Logout(token); err != nil {
		return err
	}
	return reqres.HttpOk(w, http.StatusNoContent, nil)
}

func (ac authRoute) infer(w http.ResponseWriter, r *http.Request) error {
	token := r.Header.Get("Authorization")
	if token == "" {
		return oops.Unauthorized{
			Err: errors.New("Auth token not found"),
			Msg: "Auth token not found"}
	}

	userId, err := ac.service.Infer(token)
	if err != nil {
		return err
	}

	w.Header().Add("X-User-Id", fmt.Sprintf("%d", userId))
	return reqres.HttpOk(w, http.StatusNoContent, nil)
}

func NewAuthRoute(service authService) authRoute {
	return authRoute{service}
}
