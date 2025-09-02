package internal

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/go-lib/reqres"
)

type authController struct {
	service authService
}

func NewAuthController(s authService) authController {
	return authController{s}
}

func (ac authController) Login(w http.ResponseWriter, r *http.Request) error {
	reqPayload := new(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	})

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return err
	}
	defer r.Body.Close()

	result, err := ac.service.Login(reqPayload.Username, reqPayload.Password)
	if err != nil {
		return err
	}

	resPayload := map[string]any{
		"token": map[string]any{
			"access":  result.AccessToken,
			"refresh": result.RefreshToken}}
	return reqres.HttpOk(w, http.StatusOK, resPayload)
}

func (ac authController) Register(w http.ResponseWriter, r *http.Request) error {
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

func (ac authController) Refresh(w http.ResponseWriter, r *http.Request) error {
	reqPayload := new(struct {
		Token string `json:"token"`
	})

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return err
	}
	defer r.Body.Close()

	result, err := ac.service.Refresh(reqPayload.Token)
	if err != nil {
		return err
	}

	resPayload := map[string]any{
		"token": map[string]any{
			"access":  result.AccessToken,
			"refresh": result.RefreshToken}}
	return reqres.HttpOk(w, http.StatusOK, resPayload)
}

func (ac authController) Logout(w http.ResponseWriter, r *http.Request) error {
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
