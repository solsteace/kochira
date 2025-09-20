package account

import (
	"fmt"
)

type User struct {
	Id       uint
	Username string
	Password string
	Email    string
}

func (u User) ComparePassword(
	compare func(expected, got string) error,
	password string,
) error {
	if err := compare(u.Password, password); err != nil {
		return fmt.Errorf("domain<User.MatchPassword>")
	}
	return nil
}

func NewUser(
	id *uint,
	username string,
	password string,
	email string,
) (User, error) {
	var actualId uint = 0
	if id != nil {
		actualId = *id
	}

	a := User{
		Id:       actualId,
		Username: username,
		Password: password,
		Email:    email}
	return a, nil
}
