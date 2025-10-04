package hash

type Handler interface {
	Generate(payload string) ([]byte, error)
	Compare(digest, payload string) error
}
