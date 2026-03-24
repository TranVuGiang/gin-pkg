package httpserver

import (
	"fmt"
	"strings"
)

type AppError struct {
	Code     int
	Message  string
	Internal error
}

func (e *AppError) Error() string {
	return e.Message
}

// Unwrap allows errors.Is/errors.As to unwrap the internal cause.
func (e *AppError) Unwrap() error {
	return e.Internal
}

func NewError(code int, err error, details ...string) *AppError {
	message := err.Error()

	if len(details) > 0 {
		message = fmt.Sprintf("%s: %s", message, strings.Join(details, ", "))
	}

	return &AppError{Code: code, Internal: err, Message: message}
}
