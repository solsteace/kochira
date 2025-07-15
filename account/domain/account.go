package domain

import "github.com/solsteace/kochira/internal/validation"

// Account
type Account struct {
	Id       int    `json:"id" validate:"required,min=1"`
	Email    string `json:"email" validate:"required,email,max=63"`
	Username string `json:"username" validate:"required,max=31"`
	Password string `json:"password" validate:"required,max=63"`
}

var accountValidator = validation.NewValidator()

func NewAccount(email, username, password string) (Account, error) {
	err := accountValidator.Validate(Account{Email: email, Username: username, Password: password})
	if err != nil {
		return Account{}, err
	}
	return Account{Email: email, Username: username, Password: password}, nil
}
