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
		return fmt.Errorf("internal<authService.Register>: %w", err)
	}

	user, err := domain.NewUser(nil, username, string(digest), email)
	if err != nil {
		return fmt.Errorf("internal<authService.Register>: %w", err)
	}

	if err := as.userRepo.Create(user); err != nil {
		return fmt.Errorf("internal<authService.Register>: %w", err)
	}
	return nil
}

func (as authService) Login(username, password string) (string, string, error) {
	user, err := as.userRepo.GetByUsername(username)
	if err != nil {
		return "", "", fmt.Errorf("internal<authService.Login>: %w", err)
	}

	attempts, err := as.authAttemptCache.Get(user.Id)
	if err != nil {
		return "", "", fmt.Errorf("internal<authService.Login>: %w", err)
	}
	jailTime := as.authAttemptService.CalculateJailTime(attempts)
	if jailTime > 0 {
		err := oops.Unauthorized{
			Msg: fmt.Sprintf(
				"Failed too many times! Try again in %.2fs", jailTime.Seconds())}
		return "", "", fmt.Errorf("internal<authService.Login>: %w", err)
	}

	if err := as.secret.Compare(user.Password, password); err != nil {
		attempt, _ := domain.NewAuthAttempt(false, time.Now())
		if err := as.authAttemptCache.Add(user.Id, attempt); err != nil {
			return "", "", fmt.Errorf("internal<authService.Login>: %w", err)
		}
		return "", "", fmt.Errorf("internal<authService.Login>: %w", err)
	}
	attempt, _ := domain.NewAuthAttempt(true, time.Now())
	if err := as.authAttemptCache.Add(user.Id, attempt); err != nil {
		return "", "", fmt.Errorf("internal<authService.Login>: %w", err)
	}

	accessToken, err := as.accessToken.Encode(token.NewAuth(user.Id))
	if err != nil {
		return "", "", fmt.Errorf("internal<authService.Login>: %w", err)
	}
	refreshToken, err := as.refreshToken.Encode(token.NewAuth(user.Id))
	if err != nil {
		return "", "", fmt.Errorf("internal<authService.Login>: %w", err)
	}

	if err := as.tokenCache.Grant(user.Id, refreshToken); err != nil {
		return "", "", fmt.Errorf("internal<authService.Login>: %w", err)
	}
	return accessToken, refreshToken, nil
}

func (as authService) Refresh(token string) (string, string, error) {
	payload, err := as.refreshToken.Decode(token)
	if err != nil {
		return "", "", fmt.Errorf("internal<authService.Refresh>: %w", err)
	}

	oldToken, err := as.tokenCache.FindByOwner(payload.UserId)
	switch {
	case err != nil:
		return "", "", err
	case oldToken == "":
		err := oops.Unauthorized{
			Err: errors.New("Active refresh token not found"),
			Msg: "Active refresh token not found"}
		return "", "", fmt.Errorf("internal<authService.Refresh>: %w", err)
	case oldToken != token:
		err := oops.Unauthorized{
			Err: errors.New("Given token does not match with the active one"),
			Msg: "Given token does not match with the active one"}
		return "", "", fmt.Errorf("internal<authService.Refresh>: %w", err)
	}

	accessToken, err := as.accessToken.Encode(*payload)
	if err != nil {
		return "", "", fmt.Errorf("internal<authService.Refresh>: %w", err)
	}
	refreshToken, err := as.refreshToken.Encode(*payload)
	if err != nil {
		return "", "", fmt.Errorf("internal<authService.Refresh>: %w", err)
	}

	if err := as.tokenCache.Grant(payload.UserId, refreshToken); err != nil {
		return "", "", fmt.Errorf("internal<authService.Refresh>: %w", err)
	}
	return accessToken, refreshToken, nil
}

func (as authService) Logout(token string) error {
	payload, err := as.accessToken.Decode(token)
	if err != nil {
		return fmt.Errorf("internal<authService.Logout>: %w", err)
	}

	if err := as.tokenCache.Revoke(payload.UserId); err != nil {
		return fmt.Errorf("internal<authService.Logout>: %w", err)
	}
	return nil
}

func (as authService) Infer(token string) (uint64, error) {
	payload, err := as.accessToken.Decode(token)
	if err != nil {
		return 0, fmt.Errorf("internal<authService.Infer>: %w", err)
	}

	return uint64(payload.UserId), nil
}

func (as authService) HandleNewUsers(
	maxCount uint,
	handleFx func(o []outbox.Register) ([]uint64, error),
) error {
	outbox, err := as.userRepo.GetRegisterOutbox(maxCount)
	if err != nil {
		return fmt.Errorf("internal<authService.HandleNewUsers>: %w", err)
	}
	if len(outbox) == 0 {
		return nil
	}

	sentOutbox, err := handleFx(outbox)
	if err != nil {
		return fmt.Errorf("internal<authService.HandleNewUsers>: %w", err)
	}

	if err := as.userRepo.ResolveRegisterOutbox(sentOutbox); err != nil {
		return fmt.Errorf("internal<authService.HandleNewUsers>: %w", err)
	}
	return nil
}
