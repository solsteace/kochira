package validation

import (
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validator *validator.Validate
}

// Creates API for github.com/go-playground/validator
func NewValidator() Validator {
	return Validator{
		validator: validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (vv Validator) Validate(it any) error {
	if err := vv.validator.Struct(it); err != nil {
		return err
	}
	return nil
}
