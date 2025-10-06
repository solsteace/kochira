package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/solsteace/go-lib/oops"

	"github.com/solsteace/kochira/account/internal/domain/auth"
	"github.com/solsteace/kochira/account/internal/utility/hash"
	"github.com/solsteace/kochira/account/internal/utility/token"

	authService "github.com/solsteace/kochira/account/internal/domain/auth/service"
	authAtore "github.com/solsteace/kochira/account/internal/domain/auth/store"
	authStore "github.com/solsteace/kochira/account/internal/domain/auth/store"
)

type Auth struct {
	userStore        authStore.User
	authAttemptCache authStore.Attempt
	tokenCache       authStore.Token
	refreshToken     token.Handler[token.Auth]
	accessToken      token.Handler[token.Auth]
	jailer           authService.Jailer
	hasher           hash.Handler
}

func NewAuth(
	userStore authAtore.User,
	authAttemptCache authStore.Attempt,
	tokenCache authStore.Token,
	hasher hash.Handler,
	refreshToken token.Handler[token.Auth],
	accessToken token.Handler[token.Auth],
	jailer authService.Jailer,
) Auth {
	return Auth{
		userStore:        userStore,
		authAttemptCache: authAttemptCache,
		tokenCache:       tokenCache,
		refreshToken:     refreshToken,
		accessToken:      accessToken,
		jailer:           jailer,
		hasher:           hasher}
}

func (as Auth) Login(username, password string) (string, string, error) {
	user, err := as.userStore.GetByUsername(username)
	if err != nil {
		return "", "", fmt.Errorf("service<Auth.Login>: %w", err)
	}

	attempts, err := as.authAttemptCache.Get(user.Id)
	if err != nil {
		return "", "", fmt.Errorf("service<Auth.Login>: %w", err)
	}
	if err := as.jailer.IsJailed(attempts); err != nil {
		return "", "", fmt.Errorf("service<Auth.Login>: %w", err)
	}

	authOk := true
	if err := user.ComparePassword(as.hasher.Compare, password); err != nil {
		if !errors.As(err, &oops.Unauthorized{}) {
			return "", "", fmt.Errorf("service<Auth.Login>: %w", err)
		}
		authOk = false
	}
	newAttempt, _ := auth.NewAttempt(authOk, time.Now())
	if err := as.authAttemptCache.Add(user.Id, newAttempt); err != nil {
		return "", "", fmt.Errorf("service<Auth.Login>: %w", err)
	}

	accessToken, err := as.accessToken.Encode(token.NewAuth(user.Id))
	if err != nil {
		return "", "", fmt.Errorf("service<Auth.Login>: %w", err)
	}
	refreshToken, err := as.refreshToken.Encode(token.NewAuth(user.Id))
	if err != nil {
		return "", "", fmt.Errorf("service<Auth.Login>: %w", err)
	}

	if err := as.tokenCache.Grant(user.Id, refreshToken); err != nil {
		return "", "", fmt.Errorf("service<Auth.Login>: %w", err)
	}
	return accessToken, refreshToken, nil
}

func (as Auth) Refresh(token string) (string, string, error) {
	payload, err := as.refreshToken.Decode(token)
	if err != nil {
		return "", "", fmt.Errorf("service<Auth.Refresh>: %w", err)
	}

	oldToken, err := as.tokenCache.FindByOwner(payload.UserId)
	if err != nil {
		return "", "", fmt.Errorf("service<Auth.Refresh>: %w", err)
	} else if oldToken == "" {
		return "", "", fmt.Errorf(
			"service<Auth.Refresh>: %w",
			oops.Unauthorized{
				Err: errors.New("Active refresh token not found"),
				Msg: "Active refresh token not found"})
	} else if oldToken != token {
		return "", "", fmt.Errorf(
			"service<Auth.Refresh>: %w",
			oops.Unauthorized{
				Err: errors.New("Given token does not match with the active one"),
				Msg: "Given token does not match with the active one"})
	}

	accessToken, err := as.accessToken.Encode(*payload)
	if err != nil {
		return "", "", fmt.Errorf("service<Auth.Refresh>: %w", err)
	}
	refreshToken, err := as.refreshToken.Encode(*payload)
	if err != nil {
		return "", "", fmt.Errorf("service<Auth.Refresh>: %w", err)
	}

	if err := as.tokenCache.Grant(payload.UserId, refreshToken); err != nil {
		return "", "", fmt.Errorf("service<Auth.Refresh>: %w", err)
	}
	return accessToken, refreshToken, nil
}

func (as Auth) Logout(token string) error {
	payload, err := as.accessToken.Decode(token)
	if err != nil {
		return fmt.Errorf("service<Auth.Logout>: %w", err)
	}

	if err := as.tokenCache.Revoke(payload.UserId); err != nil {
		return fmt.Errorf("service<Auth.Logout>: %w", err)
	}
	return nil
}

func (as Auth) Infer(token string) (uint64, error) {
	payload, err := as.accessToken.Decode(token)
	if err != nil {
		return 0, fmt.Errorf("service<Auth.Infer>: %w", err)
	}
	return uint64(payload.UserId), nil
}
