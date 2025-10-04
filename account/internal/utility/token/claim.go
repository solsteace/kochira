package token

type Auth struct {
	UserId uint `json:"userId"`
}

func NewAuth(userId uint) Auth {
	return Auth{UserId: userId}
}
