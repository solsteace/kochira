package domain

type User struct {
	Id       uint
	Username string
	Password string
	Email    string
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
