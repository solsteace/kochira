package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/account/internal/messaging"
	"github.com/solsteace/kochira/account/internal/service"
)

type Account struct {
	service service.Account
}

func NewAccount(service service.Account) Account {
	return Account{service: service}
}

func (s Account) Register(w http.ResponseWriter, r *http.Request) error {
	reqId := middleware.GetReqID(r.Context())
	reqPayload := new(struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	})

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return fmt.Errorf("[%s] controller<Account.Register>: %w", reqId, err)
	}
	defer r.Body.Close()

	err := s.service.Register(
		reqPayload.Username,
		reqPayload.Password,
		reqPayload.Email)
	if err != nil {
		return fmt.Errorf("[%s] controller<Account.Register>: %w", reqId, err)
	}

	resPayload := map[string]any{"msg": "Account successfully created"}
	if err := reqres.HttpOk(w, http.StatusCreated, resPayload); err != nil {
		return fmt.Errorf("[%s] controller<Account.Register>: %w", reqId, err)
	}
	return nil
}

func (a Account) PublishNewUser(maxUser uint, sender func(body []byte) error) error {
	makePayload := messaging.SerCreateSubscription
	if err := a.service.HandleRegisteredUsers(maxUser, makePayload, sender); err != nil {
		return fmt.Errorf("[%s] controller<Account.PublishNewUser>: %w", err)
	}
	return nil
}
