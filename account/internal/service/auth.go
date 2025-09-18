package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/go-lib/secret"
	"github.com/solsteace/go-lib/token"
	"github.com/solsteace/kochira/account/internal/cache"
	"github.com/solsteace/kochira/account/internal/domain"
	"github.com/solsteace/kochira/account/internal/domain/outbox"
	domainService "github.com/solsteace/kochira/account/internal/domain/service"
	"github.com/solsteace/kochira/account/internal/repository"
)

type Auth struct {
	userRepo           repository.User
	authAttemptCache   cache.AuthAttempt
	tokenCache         cache.Token
	secret             secret.Handler
	refreshToken       token.Handler[token.Auth]
	accessToken        token.Handler[token.Auth]
	authAttemptService domainService.AuthAttempt
}

func NewAuth(
	userRepo repository.User,
	authAttemptCache cache.AuthAttempt,
	tokenCache cache.Token,
	secret secret.Handler,
	refreshToken token.Handler[token.Auth],
	accessToken token.Handler[token.Auth],
	authAttemptService domainService.AuthAttempt,
) Auth {
	return Auth{
		userRepo,
		authAttemptCache,
		tokenCache,
		secret,
		refreshToken,
		accessToken,
		authAttemptService}
}

func (as Auth) Register(username, password, email string) error {
	digest, err := as.secret.Generate(password)
	if err != nil {
		return fmt.Errorf("service<Auth.Register>: %w", err)
	}

	user, err := domain.NewUser(nil, username, string(digest), email)
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
	jailTime := as.authAttemptService.CalculateJailTime(attempts)
	if jailTime > 0 {
		err := oops.Unauthorized{
			Msg: fmt.Sprintf(
				"Failed too many times! Try again in %.2fs", jailTime.Seconds())}
		return "", "", fmt.Errorf("service<Auth.Login>: %w", err)
	}

	if err := as.secret.Compare(user.Password, password); err != nil {
		attempt, _ := domain.NewAuthAttempt(false, time.Now())
		if err := as.authAttemptCache.Add(user.Id, attempt); err != nil {
			return "", "", fmt.Errorf("service<Auth.Login>: %w", err)
		}
		return "", "", fmt.Errorf("service<Auth.Login>: %w", err)
	}
	attempt, _ := domain.NewAuthAttempt(true, time.Now())
	if err := as.authAttemptCache.Add(user.Id, attempt); err != nil {
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

func (as Auth) HandleNewUsers(
	maxCount uint,
	handleFx func(o []outbox.Register) ([]uint64, error),
) error {
	outbox, err := as.userRepo.GetRegisterOutbox(maxCount)
	if err != nil {
		return fmt.Errorf("service<Auth.HandleNewUsers>: %w", err)
	}
	if len(outbox) == 0 {
		return nil
	}

	sentOutbox, err := handleFx(outbox)
	if err != nil {
		return fmt.Errorf("service<Auth.HandleNewUsers>: %w", err)
	}

	if err := as.userRepo.ResolveRegisterOutbox(sentOutbox); err != nil {
		return fmt.Errorf("service<Auth.HandleNewUsers>: %w", err)
	}
	return nil
}
