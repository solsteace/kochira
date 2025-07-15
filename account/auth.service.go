package account

import (
	"fmt"

	"github.com/solsteace/kochira/account/repository"
)

type AuthService struct {
	accountRepo repository.AccountRepo
}

func (as AuthService) login(username, password string) (string, error) {
	account, err := as.accountRepo.FindById(1)
	if err != nil {
		return "", nil
	}

	return fmt.Sprintf("%+v", account), nil
}

func (as AuthService) register(username, password, email string) (string, error) {
	return "You're registering!", nil
}
