package account

import (
	"net/http"

	"github.com/solsteace/kochira/internal/validation"
)

type AuthController struct {
	service AuthService
}

// Content-type: application/www-x-form-urlencoded
func (ac AuthController) login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	payload := struct {
		Username string `validate:"required"`
		Password string `validate:"required"`
	}{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}
	validator := validation.NewValidator()
	if err := validator.Validate(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	msg, err := ac.service.login(payload.Username, payload.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msg))
}

// Content-type: application/www-x-form-urlencoded
func (ac AuthController) register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	payload := struct {
		Username string `validate:"required"`
		Password string `validate:"required"`
		Email    string `validate:"required"`
	}{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		Email:    r.FormValue("email"),
	}
	validator := validation.NewValidator()
	if err := validator.Validate(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	msg, err := ac.service.register(
		payload.Username, payload.Password, payload.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msg))
}
