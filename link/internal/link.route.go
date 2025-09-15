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

type linkRoute struct {
	service     linkService
	userContext middleware.UserContext
}

func (lr linkRoute) Use(parent *chi.Mux) {
	link := chi.NewRouter()
	link.Group(func(r chi.Router) {
		r.Use(lr.userContext.Handle)
		r.Get("/my", reqres.HttpHandlerWithError(lr.getSelf))
		r.Get("/my/{id}", reqres.HttpHandlerWithError(lr.getById))
		r.Post("/", reqres.HttpHandlerWithError(lr.create))
		r.Put("/{id}", reqres.HttpHandlerWithError(lr.updateById))
		r.Delete("/{id}", reqres.HttpHandlerWithError(lr.deleteById))
	})
	parent.Mount("/link", link)
}

func (lr linkRoute) getSelf(w http.ResponseWriter, r *http.Request) error {
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
	result, err := lr.service.GetSelf(uint64(userId), page, limit)
	if err != nil {
		return err
	}
	return reqres.HttpOk(w, http.StatusOK, result)
}

func (lr linkRoute) getById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return err
	}

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	result, err := lr.service.GetById(uint64(userId), id)
	if err != nil {
		return err
	}
	return reqres.HttpOk(w, http.StatusOK, result)
}

func (lr linkRoute) create(w http.ResponseWriter, r *http.Request) error {
	reqPayload := new(struct {
		Destination string `json:"destination"`
	})
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(reqPayload); err != nil {
		return err
	}
	defer r.Body.Close()

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	err := lr.service.Create(uint64(userId), reqPayload.Destination)
	if err != nil {
		return err
	}
	return reqres.HttpOk(w, http.StatusCreated, nil)
}

func (lr linkRoute) updateById(w http.ResponseWriter, r *http.Request) error {
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
	err = lr.service.UpdateById(
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

func (lr linkRoute) deleteById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return err
	}

	userId := r.Context().Value(middleware.UserContextCtxKey).(middleware.UserContextCtxPayload)
	if err := lr.service.DeleteById(uint64(userId), id); err != nil {
		return err
	}
	return reqres.HttpOk(w, http.StatusNoContent, nil)
}

func NewLinkRoute(service linkService, userContext middleware.UserContext) linkRoute {
	return linkRoute{service, userContext}
}
