package validator

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	Validator *validator.Validate
}

var (
	once             sync.Once
	defaultValidator *Validator
)

func DefaultRestValidator() *Validator {
	once.Do(func() {
		defaultValidator = &Validator{Validator: validator.New()}
	})

	return defaultValidator
}

func (v *Validator) Validate(i any) error {
	if err := v.Validator.Struct(i); err != nil {
		return err
	}

	return nil
}
