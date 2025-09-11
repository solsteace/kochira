package internal

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

type authService struct {
	userRepo           repository.User
	authAttemptCache   cache.AuthAttempt
	tokenCache         cache.Token
	secret             secret.Handler
	refreshToken       token.Handler[token.Auth]
	accessToken        token.Handler[token.Auth]
	authAttemptService domainService.AuthAttempt
}

func NewAuthService(
	userRepo repository.User,
	authAttemptCache cache.AuthAttempt,
	tokenCache cache.Token,
	secret secret.Handler,
	refreshToken token.Handler[token.Auth],
	accessToken token.Handler[token.Auth],
	authAttemptService domainService.AuthAttempt,
) authService {
	return authService{
		userRepo,
		authAttemptCache,
		tokenCache,
		secret,
		refreshToken,
		accessToken,
		authAttemptService}
}

func (as authService) Register(username, password, email string) error {
	digest, err := as.secret.Generate(password)
	if err != nil {
		return err
	}

	user, err := domain.NewUser(nil, username, string(digest), email)
	if err != nil {
		return err
	}

	if err := as.userRepo.Create(user); err != nil {
		return err
	}
	return nil
}

func (as authService) Login(username, password string) (string, string, error) {
	user, err := as.userRepo.GetByUsername(username)
	if err != nil {
		return "", "", err
	}

	attempts, err := as.authAttemptCache.Get(user.Id)
	if err != nil {
		return "", "", err
	}
	jailTime := as.authAttemptService.CalculateJailTime(attempts)
	if jailTime > 0 {
		return "", "", oops.Unauthorized{
			Msg: fmt.Sprintf(
				"Failed too many times! Try again in %.2fs", jailTime.Seconds())}
	}

	if err := as.secret.Compare(user.Password, password); err != nil {
		attempt, _ := domain.NewAuthAttempt(false, time.Now())
		if err := as.authAttemptCache.Add(user.Id, attempt); err != nil {
			return "", "", err
		}
		return "", "", err
	}
	attempt, _ := domain.NewAuthAttempt(true, time.Now())
	if err := as.authAttemptCache.Add(user.Id, attempt); err != nil {
		return "", "", err
	}

	accessToken, err := as.accessToken.Encode(token.NewAuth(user.Id))
	if err != nil {
		return "", "", err
	}
	refreshToken, err := as.refreshToken.Encode(token.NewAuth(user.Id))
	if err != nil {
		return "", "", err
	}

	if err := as.tokenCache.Grant(user.Id, refreshToken); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (as authService) Refresh(token string) (string, string, error) {
	payload, err := as.refreshToken.Decode(token)
	if err != nil {
		return "", "", err
	}

	oldToken, err := as.tokenCache.FindByOwner(payload.UserId)
	switch {
	case err != nil:
		return "", "", err
	case oldToken == "":
		return "", "", oops.Unauthorized{
			Err: errors.New("Active refresh token not found"),
			Msg: "Active refresh token not found"}
	case oldToken != token:
		return "", "", oops.Unauthorized{
			Err: errors.New("Given token does not match with the active one"),
			Msg: "Given token does not match with the active one"}
	}

	accessToken, err := as.accessToken.Encode(*payload)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := as.refreshToken.Encode(*payload)
	if err != nil {
		return "", "", err
	}

	if err := as.tokenCache.Grant(payload.UserId, refreshToken); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (as authService) Logout(token string) error {
	payload, err := as.accessToken.Decode(token)
	if err != nil {
		return err
	}

	return as.tokenCache.Revoke(payload.UserId)
}

func (as authService) Infer(token string) (uint64, error) {
	payload, err := as.accessToken.Decode(token)
	if err != nil {
		return 0, err
	}

	return uint64(payload.UserId), nil
}

func (as authService) HandleNewUsers(
	maxCount uint,
	handleFx func(o []outbox.Register) ([]uint64, error),
) error {
	outbox, err := as.userRepo.GetRegisterOutbox(maxCount)
	if err != nil {
		return err
	}
	if len(outbox) == 0 {
		return nil
	}

	sentOutbox, err := handleFx(outbox)
	if err != nil {
		return err
	}
	return as.userRepo.ResolveRegisterOutbox(sentOutbox)
}
