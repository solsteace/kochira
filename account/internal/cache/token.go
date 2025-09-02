package cache

type Token interface {
	Grant(userId uint, token string) error
	Revoke(userId uint) error
	FindByOwner(userId uint) (string, error)
}
