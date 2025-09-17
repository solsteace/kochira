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

func HttpErrorMsg(err error) string {
	var lastErr error
	for err != nil {
		lastErr = err
		err = errors.Unwrap(err)
	}

	switch {
	case errors.As(lastErr, &oops.BadRequest{}),
		errors.As(lastErr, &oops.BadValues{}),
		errors.As(lastErr, &oops.Unauthorized{}),
		errors.As(lastErr, &oops.Forbidden{}),
		errors.As(lastErr, &oops.NotFound{}):
		return lastErr.Error()
	}
	return "internal server error"
}
