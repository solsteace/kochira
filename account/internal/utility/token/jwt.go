package token

import (
	"errors"
	"fmt"
	"time"

	goJwt "github.com/golang-jwt/jwt/v5"
	"github.com/solsteace/go-lib/oops"
)

type jwt[PayloadType any] struct {
	method   goJwt.SigningMethod
	secret   []byte
	lifetime time.Duration
	issuer   string
}

// Creates jwt token handler for jwt with `PayloadType` payload. The token would be
// signed with `secret` and valid for `lifetime` seconds after its creation.
func NewJwt[PayloadType any](
	issuer,
	secret string,
	lifetime time.Duration,
) jwt[PayloadType] {
	return jwt[PayloadType]{
		method:   goJwt.SigningMethodHS256, // Make it customizable later
		secret:   []byte(secret),
		lifetime: lifetime,
		issuer:   issuer,
	}
}

// TODO: change. I couldn't think of a better name for now
type theClaims[PayloadType any] struct {
	Payload PayloadType
	goJwt.RegisteredClaims
}

func (j jwt[PayloadType]) Encode(payload PayloadType) (string, error) {
	now := time.Now()
	claims := theClaims[PayloadType]{
		Payload: payload,
		RegisteredClaims: goJwt.RegisteredClaims{
			ExpiresAt: goJwt.NewNumericDate(now.Add(j.lifetime * time.Second)),
			IssuedAt:  goJwt.NewNumericDate(now),
			NotBefore: goJwt.NewNumericDate(now),
			Issuer:    j.issuer,
			// Could be added more, but this'll be enough for now
		},
	}

	token := goJwt.NewWithClaims(j.method, claims)
	signedToken, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("utility<jwt.Encode>: %w", err)
	}
	return signedToken, nil
}

func (j jwt[PayloadType]) Decode(token string) (*PayloadType, error) {
	// https://pkg.go.dev/github.com/golang-jwt/jwt/v5#example-ParseWithClaims-CustomClaimsType
	parsedToken, err := goJwt.ParseWithClaims(
		token,
		new(theClaims[PayloadType]),
		func(t *goJwt.Token) (any, error) {
			return j.secret, nil
		},
		goJwt.WithValidMethods([]string{j.method.Alg()}))
	if err != nil {
		switch {
		case errors.Is(err, goJwt.ErrTokenExpired):
			err2 := oops.Unauthorized{
				Err: err,
				Msg: "Token has been expired"}
			return nil, fmt.Errorf("utility<jwt.Decode>: %w", err2)
		case errors.Is(err, goJwt.ErrTokenUsedBeforeIssued):
			err2 := oops.Unauthorized{
				Err: err,
				Msg: "Token is used ahead of its time"}
			return nil, fmt.Errorf("utility<jwt.Decode>: %w", err2)
		case errors.Is(err, goJwt.ErrTokenMalformed):
			err2 := oops.Unauthorized{
				Err: err,
				Msg: "Token is invalid"}
			return nil, fmt.Errorf("utility<jwt.Decode>: %w", err2)
		default:
			return nil, fmt.Errorf("utility<jwt.Decode>: %w", err)
		}
	}

	claims, ok := parsedToken.Claims.(*theClaims[PayloadType])
	if !ok {
		err := errors.New("Claims type is not `theClaims[PayloadType]`")
		return nil, fmt.Errorf("utility<jwt.Decode>: %w", err)
	}
	return &claims.Payload, nil
}
