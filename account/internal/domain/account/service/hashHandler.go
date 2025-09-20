package service

type HashHandler interface {
	Generate(payload string) ([]byte, error)
	Compare(digest, payload string) error
}
