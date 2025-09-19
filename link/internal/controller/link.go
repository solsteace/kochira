package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/link/internal/middleware"
	"github.com/solsteace/kochira/link/internal/service"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type Link struct {
	service service.Link
}

func (lr Link) GetSelf(w http.ResponseWriter, r *http.Request) error {
	reqId := chiMiddleware.GetReqID(r.Context())
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

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	result, err := lr.service.GetSelf(uint64(userId), page, limit)
	if err != nil {
		return fmt.Errorf("[%s] controller<Link.GetSelf>: %w", reqId, err)
	}

	if err := reqres.HttpOk(w, http.StatusOK, result); err != nil {
		return fmt.Errorf("[%s] controller<Link.GetSelf>: %w", reqId, err)
	}
	return nil
}

func (lr Link) GetById(w http.ResponseWriter, r *http.Request) error {
	reqId := chiMiddleware.GetReqID(r.Context())
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return fmt.Errorf("[%s] controller<Link.GetById>: %w", reqId, err)
	}

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	result, err := lr.service.GetById(uint64(userId), id)
	if err != nil {
		return fmt.Errorf("[%s] controller<Link.GetById>: %w", reqId, err)
	}

	if err := reqres.HttpOk(w, http.StatusOK, result); err != nil {
		return fmt.Errorf("[%s] controller<Link.GetById>: %w", reqId, err)
	}
	return nil
}

func (lr Link) Create(w http.ResponseWriter, r *http.Request) error {
	reqId := chiMiddleware.GetReqID(r.Context())
	reqPayload := new(struct {
		Destination string `json:"destination"`
	})
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return fmt.Errorf("[%s] controller<Link.Create>: %w", reqId, err)
	}
	defer r.Body.Close()

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	err := lr.service.Create(uint64(userId), reqPayload.Destination)
	if err != nil {
		return fmt.Errorf("[%s] controller<Link.Create>: %w", reqId, err)
	}

	if err := reqres.HttpOk(w, http.StatusCreated, nil); err != nil {
		return fmt.Errorf("[%s] controller<Link.Create>: %w", reqId, err)
	}
	return nil
}

func (lr Link) UpdateById(w http.ResponseWriter, r *http.Request) error {
	reqId := chiMiddleware.GetReqID(r.Context())
	reqPayload := new(struct {
		Shortened   string `json:"shortened"`
		Destination string `json:"destination"`
		IsOpen      bool   `json:"is_open"`
	})
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return fmt.Errorf("[%s] controller<Link.UpdateById>: %w", reqId, err)
	}
	defer r.Body.Close()

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return fmt.Errorf("[%s] controller<Link.UpdateById>: %w", reqId, err)
	}

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	err = lr.service.UpdateById(
		uint64(userId),
		id,
		reqPayload.Shortened,
		reqPayload.Destination,
		reqPayload.IsOpen)
	if err != nil {
		return fmt.Errorf("[%s] controller<Link.UpdateById>: %w", reqId, err)
	}

	if err := reqres.HttpOk(w, http.StatusOK, nil); err != nil {
		return fmt.Errorf("[%s] controller<Link.Create>: %w", reqId, err)
	}
	return nil
}

func (lr Link) DeleteById(w http.ResponseWriter, r *http.Request) error {
	reqId := chiMiddleware.GetReqID(r.Context())
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return fmt.Errorf("[%s] controller<Link.DeleteById>: %w", reqId, err)
	}

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	if err := lr.service.DeleteById(uint64(userId), id); err != nil {
		return fmt.Errorf("[%s] controller<Link.DeleteById>: %w", reqId, err)
	}

	if err := reqres.HttpOk(w, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("[%s] controller<Link.Create>: %w", reqId, err)
	}
	return nil
}

func (l Link) Redirect(w http.ResponseWriter, r *http.Request) error {
	reqId := chiMiddleware.GetReqID(r.Context())
	shortened := chi.URLParam(r, "shortened")
	destination, err := l.service.Redirect(shortened)
	if err != nil {
		return fmt.Errorf("[%s] controller<Redirection.Redirect>: %w", reqId, err)
	}

	http.Redirect(w, r, destination, http.StatusTemporaryRedirect)
	return nil
}

func NewLink(service service.Link) Link {
	return Link{service}
}
