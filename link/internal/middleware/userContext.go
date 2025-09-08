package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/go-lib/reqres"
)

const UserContextCtxKey = "userId"

type UserContextCtxPayload uint64

type UserContext struct {
	header string
}

func NewUserContext(header string) UserContext {
	return UserContext{header}
}

func (uc UserContext) Handle(next http.Handler) http.Handler {
	return reqres.HttpHandlerWithError(
		func(w http.ResponseWriter, r *http.Request) error {
			userCtx := r.Context()

			xUserId := r.Header.Get(uc.header)
			userId, err := strconv.ParseUint(xUserId, 10, 64)
			if err != nil {
				return oops.BadRequest{
					Err: err,
					Msg: "Something went wrong during obtaining user context"}
			}

			var payload UserContextCtxPayload = UserContextCtxPayload(userId)
			userCtx = context.WithValue(userCtx, UserContextCtxKey, payload)
			next.ServeHTTP(w, r.WithContext(userCtx))
			return nil
		})
}
