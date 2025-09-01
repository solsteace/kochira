package account

import (
	"fmt"
	"time"

	"github.com/solsteace/go-lib/oops"
	"github.com/solsteace/go-lib/secret"
	"github.com/solsteace/go-lib/token"
	"github.com/solsteace/kochira/internal/account/cache"
	"github.com/solsteace/kochira/internal/account/domain"
	domainService "github.com/solsteace/kochira/internal/account/domain/service"
	"github.com/solsteace/kochira/internal/account/repository"
)

type authService struct {
	userRepo           repository.Account
	authAttemptCache   cache.AuthAttempt
	secret             secret.Handler
	refreshToken       token.Handler[token.Auth]
	accessToken        token.Handler[token.Auth]
	authAttemptService domainService.AuthAttempt
}

func NewAuthService(
	userRepo repository.Account,
	authAttemptCache cache.AuthAttempt,
	secret secret.Handler,
	refreshToken token.Handler[token.Auth],
	accessToken token.Handler[token.Auth],
	authAttemptService domainService.AuthAttempt,
) authService {
	return authService{
		userRepo,
		authAttemptCache,
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
	return result, nil
}
