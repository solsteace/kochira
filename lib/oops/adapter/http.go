package adapter

import (
	"errors"
	"net/http"

	"github.com/solsteace/go-lib/oops"
)

func HttpStatusCode(err error) int {
	switch {
	case errors.As(err, &oops.BadRequest{}),
		errors.As(err, &oops.BadValues{}):
		return http.StatusBadRequest
	case errors.As(err, &oops.Unauthorized{}):
		return http.StatusUnauthorized
	case errors.As(err, &oops.Forbidden{}):
		return http.StatusForbidden
	case errors.As(err, &oops.NotFound{}):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
