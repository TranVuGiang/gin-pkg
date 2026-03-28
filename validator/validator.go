package validator

import (
	"reflect"
	"strings"
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
		v := validator.New()
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" || name == "" {
				return fld.Name
			}
			return name
		})
		defaultValidator = &Validator{Validator: v}
	})

	return defaultValidator
}

func (v *Validator) Validate(i any) error {
	return v.Validator.Struct(i)
}
