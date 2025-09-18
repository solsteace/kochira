package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/account/internal/service"
)

type Auth struct {
	service service.Auth
}

func (s Auth) Login(w http.ResponseWriter, r *http.Request) error {
	reqId := middleware.GetReqID(r.Context())
	reqPayload := new(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	})

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return fmt.Errorf("[%s] controller<Auth.Login>: %w", reqId, err)
	}
	defer r.Body.Close()

	accessToken, refreshToken, err := s.service.Login(reqPayload.Username, reqPayload.Password)
	if err != nil {
		return fmt.Errorf("[%s] controller<Auth.Login>: %w", reqId, err)
	}

	resPayload := map[string]any{
		"token": map[string]any{
			"access":  accessToken,
			"refresh": refreshToken}}
	if err := reqres.HttpOk(w, http.StatusOK, resPayload); err != nil {
		return fmt.Errorf("[%s] controller<Auth.Login>: %w", reqId, err)
	}
	return nil
}

func (s Auth) Register(w http.ResponseWriter, r *http.Request) error {
	reqId := middleware.GetReqID(r.Context())
	reqPayload := new(struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	})

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return fmt.Errorf("[%s] controller<Auth.Register>: %w", reqId, err)
	}
	defer r.Body.Close()

	err := s.service.Register(
		reqPayload.Username,
		reqPayload.Password,
		reqPayload.Email)
	if err != nil {
		return fmt.Errorf("[%s] controller<Auth.Register>: %w", reqId, err)
	}

	resPayload := map[string]any{"msg": "Account successfully created"}
	if err := reqres.HttpOk(w, http.StatusCreated, resPayload); err != nil {
		return fmt.Errorf("[%s] controller<Auth.Register>: %w", reqId, err)
	}
	return nil
}

func (s Auth) Refresh(w http.ResponseWriter, r *http.Request) error {
	reqId := middleware.GetReqID(r.Context())
	reqPayload := new(struct {
		Token string `json:"token"`
	})

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return fmt.Errorf("[%s] controller<Auth.Refresh>: %w", reqId, err)
	}
	defer r.Body.Close()

	accessToken, refreshToken, err := s.service.Refresh(reqPayload.Token)
	if err != nil {
		return fmt.Errorf("[%s] controller<Auth.Refresh>: %w", reqId, err)
	}

	resPayload := map[string]any{
		"token": map[string]any{
			"access":  accessToken,
			"refresh": refreshToken}}
	if err := reqres.HttpOk(w, http.StatusOK, resPayload); err != nil {
		return fmt.Errorf("[%s] controller<Auth.Refresh>: %w", reqId, err)
	}
	return nil
}

func (a Auth) Logout(w http.ResponseWriter, r *http.Request) error {
	reqId := middleware.GetReqID(r.Context())
	token := r.Header.Get("Authorization")
	if token == "" {
		err := oops.Unauthorized{
			Err: errors.New("Auth token not found"),
			Msg: "Auth token not found"}
		return fmt.Errorf("[%s] controller<Auth.Logout>: %w", reqId, err)
	}

	if err := a.service.Logout(token); err != nil {
		return fmt.Errorf("[%s] controller<Auth.Logout>: %w", reqId, err)
	}

	if err := reqres.HttpOk(w, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("[%s] controller<Auth.Logout>: %w", reqId, err)
	}
	return nil
}

func (a Auth) Infer(w http.ResponseWriter, r *http.Request) error {
	reqId := middleware.GetReqID(r.Context())
	token := r.Header.Get("Authorization")
	if token == "" {
		err := oops.Unauthorized{
			Err: errors.New("Auth token not found"),
			Msg: "Auth token not found"}
		return fmt.Errorf("[%s] controller<Auth.Infer>: %w", reqId, err)
	}

	userId, err := a.service.Infer(token)
	if err != nil {
		return fmt.Errorf("[%s] controller<Auth.Infer>: %w", reqId, err)
	}

	w.Header().Add("X-User-Id", fmt.Sprintf("%d", userId))
	if err := reqres.HttpOk(w, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("[%s] controller<Auth.Infer>: %w", reqId, err)
	}
	return nil
}

func NewAuth(service service.Auth) Auth {
	return Auth{service}
}
