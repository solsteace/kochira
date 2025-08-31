package token

type Handler[PayloadType any] interface {
	Encode(payload PayloadType) (string, error)
	Decode(token string) (*PayloadType, error)
}
