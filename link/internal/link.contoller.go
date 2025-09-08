package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/link/internal/middleware"
)

type linkController struct {
	service linkService
}

func NewLinkController(service linkService) linkController {
	return linkController{service}
}

func (lc linkController) GetSelf(w http.ResponseWriter, r *http.Request) error {
	var limit *uint
	var page *uint
	rq := r.URL.Query()
	qLimit, err := strconv.ParseUint(rq.Get("limit"), 10, 64)
	if err == nil {
		temp := uint(qLimit)
		limit = &temp
	}
	qPage, err := strconv.ParseUint(rq.Get("page"), 10, 64)
	if err == nil {
		temp := uint(qPage)
		page = &temp
	}

	fmt.Println(*limit, *page)
	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	result, err := lc.service.GetSelf(uint64(userId), page, limit)
	if err != nil {
		return err
	}
	return reqres.HttpOk(w, http.StatusOK, result)
}

func (lc linkController) GetById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return err
	}

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	result, err := lc.service.GetById(uint64(userId), id)
	if err != nil {
		return err
	}
	return reqres.HttpOk(w, http.StatusOK, result)
}

func (lc linkController) Create(w http.ResponseWriter, r *http.Request) error {
	reqPayload := new(struct {
		Destination string `json:"destination"`
	})
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return err
	}
	defer r.Body.Close()

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	err := lc.service.Create(uint64(userId), reqPayload.Destination)
	if err != nil {
		return err
	}
	return reqres.HttpOk(w, http.StatusCreated, nil)
}

func (lc linkController) UpdateById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return err
	}

	reqPayload := new(struct {
		Shortened   string `json:"shortened"`
		Destination string `json:"destination"`
		IsOpen      bool   `json:"is_open"`
	})
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return err
	}
	defer r.Body.Close()

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	err = lc.service.UpdateById(
		uint64(userId),
		id,
		reqPayload.Shortened,
		reqPayload.Destination,
		reqPayload.IsOpen)
	if err != nil {
		return err
	}
	return reqres.HttpOk(w, http.StatusOK, nil)
}

func (lc linkController) DeleteById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return err
	}

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	if err := lc.service.DeleteById(uint64(userId), id); err != nil {
		return err
	}
	return reqres.HttpOk(w, http.StatusNoContent, nil)
}

// ==================================================
// Non-CRUD
// ==================================================

func (lc linkController) Redirect(w http.ResponseWriter, r *http.Request) error {
	shortened := chi.URLParam(r, "shortened")
	destination, err := lc.service.Redirect(shortened)
	if err != nil {
		return err
	}

	http.Redirect(w, r, destination, http.StatusTemporaryRedirect)
	return nil
}
