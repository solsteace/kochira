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

func (as authService) Login(username, password string) (
	*struct {
		RefreshToken string
		AccessToken  string
	},
	error,
) {
	result := new(struct {
		RefreshToken string
		AccessToken  string
	})

	user, err := as.userRepo.GetByUsername(username)
	if err != nil {
		return result, err
	}

	attempts, err := as.authAttemptCache.Get(user.Id)
	if err != nil {
		return result, err
	}
	jailTime := as.authAttemptService.CalculateJailTime(attempts)
	if jailTime > 0 {
		return result, oops.Unauthorized{
			Msg: fmt.Sprintf(
				"Failed too many times! Try again in %.2fs", jailTime.Seconds())}
	}

	if err := as.secret.Compare(user.Password, password); err != nil {
		attempt, _ := domain.NewAuthAttempt(false, time.Now())
		if err := as.authAttemptCache.Add(user.Id, attempt); err != nil {
			return result, err
		}
		return result, err
	}
	attempt, _ := domain.NewAuthAttempt(true, time.Now())
	if err := as.authAttemptCache.Add(user.Id, attempt); err != nil {
		return result, err
	}

	accessToken, err := as.accessToken.Encode(token.NewAuth(user.Id))
	if err != nil {
		return result, err
	}
	refreshToken, err := as.refreshToken.Encode(token.NewAuth(user.Id))
	if err != nil {
		return result, err
	}

	if err := as.tokenCache.Grant(user.Id, refreshToken); err != nil {
		return result, err
	}

	result.RefreshToken = refreshToken
	result.AccessToken = accessToken
	return result, nil
}

func (as authService) Refresh(token string) (
	*struct {
		RefreshToken string
		AccessToken  string
	},
	error,
) {
	result := new(struct {
		RefreshToken string
		AccessToken  string
	})

	payload, err := as.refreshToken.Decode(token)
	if err != nil {
		return result, err
	}

	oldToken, err := as.tokenCache.FindByOwner(payload.UserId)
	switch {
	case err != nil:
		return result, err
	case oldToken == "":
		return result, oops.Unauthorized{
			Err: errors.New("Active refresh token not found"),
			Msg: "Active refresh token not found"}
	case oldToken != token:
		return result, oops.Unauthorized{
			Err: errors.New("Given token does not match with the active one"),
			Msg: "Given token does not match with the active one"}
	}

	accessToken, err := as.accessToken.Encode(*payload)
	if err != nil {
		return result, err
	}
	refreshToken, err := as.refreshToken.Encode(*payload)
	if err != nil {
		return result, err
	}

	if err := as.tokenCache.Grant(payload.UserId, refreshToken); err != nil {
		return result, err
	}

	result.AccessToken = accessToken
	result.RefreshToken = refreshToken
	return result, nil
}

func (as authService) Logout(token string) error {
	payload, err := as.accessToken.Decode(token)
	if err != nil {
		return err
	}

	return as.tokenCache.Revoke(payload.UserId)
}

func (as authService) Infer(token string) (
	*struct{ UserId uint },
	error,
) {
	result := new(struct{ UserId uint })

	payload, err := as.accessToken.Decode(token)
	if err != nil {
		return result, err
	}

	result.UserId = payload.UserId
	return result, nil
}
