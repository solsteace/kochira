package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/go-lib/token"

	"github.com/solsteace/kochira/account/internal/domain/account"
	"github.com/solsteace/kochira/account/internal/domain/auth"

	accountService "github.com/solsteace/kochira/account/internal/domain/account/service"
	accountStore "github.com/solsteace/kochira/account/internal/domain/account/store"
	authService "github.com/solsteace/kochira/account/internal/domain/auth/service"
	authStore "github.com/solsteace/kochira/account/internal/domain/auth/store"
)

type Auth struct {
	userRepo         accountStore.User
	authAttemptCache authStore.Attempt
	tokenCache       authStore.Token
	refreshToken     token.Handler[token.Auth]
	accessToken      token.Handler[token.Auth]
	jailer           authService.Jailer
	hashHandler      accountService.HashHandler
}

func NewAuth(
	userRepo accountStore.User,
	authAttemptCache authStore.Attempt,
	tokenCache authStore.Token,
	hasher accountService.HashHandler,
	refreshToken token.Handler[token.Auth],
	accessToken token.Handler[token.Auth],
	authJailer authService.Jailer,
) Auth {
	return Auth{
		userRepo,
		authAttemptCache,
		tokenCache,
		refreshToken,
		accessToken,
		authJailer,
		hasher}
}

func (as Auth) Register(username, password, email string) error {
	digest, err := as.hashHandler.Generate(password)
	if err != nil {
		return fmt.Errorf("service<Auth.Register>: %w", err)
	}

	user, err := account.NewUser(nil, username, string(digest), email)
	if err != nil {
		return fmt.Errorf("service<Auth.Register>: %w", err)
	}

	if err := as.userRepo.Create(user); err != nil {
		return fmt.Errorf("service<Auth.Register>: %w", err)
	}
	return nil
}

func (as Auth) Login(username, password string) (string, string, error) {
	user, err := as.userRepo.GetByUsername(username)
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
	if err := user.ComparePassword(as.hashHandler.Compare, password); err != nil {
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
	switch {
	case err != nil:
		return "", "", err
	case oldToken == "":
		err := oops.Unauthorized{
			Err: errors.New("Active refresh token not found"),
			Msg: "Active refresh token not found"}
		return "", "", fmt.Errorf("service<Auth.Refresh>: %w", err)
	case oldToken != token:
		err := oops.Unauthorized{
			Err: errors.New("Given token does not match with the active one"),
			Msg: "Given token does not match with the active one"}
		return "", "", fmt.Errorf("service<Auth.Refresh>: %w", err)
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

func (as Auth) HandleRegisteredUsers(
	maxCount uint,
	payloader func(users []uint64) ([]byte, error),
	send func([]byte) error,
) error {
	outbox, err := as.userRepo.GetRegisterOutbox(maxCount)
	switch {
	case err != nil:
		return fmt.Errorf("service<Auth.HandleNewUsers>: %w", err)
	case len(outbox) == 0:
		return nil
	}

	handledUser := []uint64{}
	for _, o := range outbox {
		handledUser = append(handledUser, o.UserId())
	}
	payload, err := payloader(handledUser)
	if err != nil {
		return fmt.Errorf("service<Auth.HandleNewUsers>: %w", err)
	}
	if err := send(payload); err != nil {
		return fmt.Errorf("service<Auth.HandleNewUsers>: %w", err)
	}

	if err := as.userRepo.ResolveRegisterOutbox(handledUser); err != nil {
		return fmt.Errorf("service<Auth.HandleNewUsers>: %w", err)
	}
	return nil
}
