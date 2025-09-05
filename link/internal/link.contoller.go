package internal

import (
	"encoding/json"
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
	ctx := r.Context()
	userId := ctx.Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)

	rq := r.URL.Query()
	limit, err := strconv.ParseUint(rq.Get("limit"), 10, 64)
	if err != nil {
		return err
	}
	page, err := strconv.ParseUint(rq.Get("page"), 10, 64)
	if err != nil {
		return err
	}

	result, err := lc.service.GetSelf(uint(userId), uint(page), uint(limit))
	if err != nil {
		return err
	}
	return reqres.HttpOk(w, http.StatusOK, result)
}

func (lc linkController) GetById(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userId := ctx.Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return err
	}

	result, err := lc.service.GetById(uint(userId), uint(id))
	if err != nil {
		return err
	}
	return reqres.HttpOk(w, http.StatusOK, result)
}

func (lc linkController) Create(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userId := ctx.Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)

	reqPayload := new(struct {
		Destination string `json:"destination"`
	})
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return err
	}
	defer r.Body.Close()

	err := lc.service.Create(
		uint(userId),
		reqPayload.Destination)
	if err != nil {
		return err
	}

	return nil
}

func (lc linkController) UpdateById(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userId := ctx.Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return err
	}

	reqPayload := new(struct {
		Destination string `json:"destination"`
		Shortened   string `json:"shortened"`
		IsOpen      bool   `json:"is_open"`
	})
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return err
	}
	defer r.Body.Close()

	if err := lc.service.UpdateById(uint(userId), uint(id)); err != nil {
		return err
	}
	return nil
}

func (lc linkController) DeleteById(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userId := ctx.Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return err
	}

	if err := lc.service.Delete(uint(userId), uint(id)); err != nil {
		return err
	}
	return nil
}

// ==================================================
// Non-CRUD
// ==================================================

func (lc linkController) Redirect(w http.ResponseWriter, r *http.Request) error {
	return nil
}
