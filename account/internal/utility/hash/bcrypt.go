package hash

import (
	"errors"
	"fmt"

	"github.com/solsteace/go-lib/oops"
	bc "golang.org/x/crypto/bcrypt"
)

type bcrypt struct {
	cost int
}

func NewBcrypt(cost int) bcrypt {
	return bcrypt{cost: cost}
}

func (b bcrypt) Generate(payload string) ([]byte, error) {
	return bc.GenerateFromPassword([]byte(payload), b.cost)
}

func (b bcrypt) Compare(digest, payload string) error {
	switch err := bc.CompareHashAndPassword([]byte(digest), []byte(payload)); {
	case errors.Is(err, bc.ErrMismatchedHashAndPassword):
		err2 := oops.Unauthorized{
			Err: err,
			Msg: "Password doesn't match"}
		return fmt.Errorf("utility<bcrypt.Compare>: %w", err2)
	case err != nil:
		return fmt.Errorf("utility<bcrypt.Compare>: %w", err)
	}
	return nil
}
