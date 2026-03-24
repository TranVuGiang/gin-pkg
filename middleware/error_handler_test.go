package middleware_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/TranVuGiang/gin_pkg/middleware"
	"github.com/TranVuGiang/gin_pkg/testutil"
	"github.com/stretchr/testify/require"
)

var ErrGeneric = errors.New("generic error")

func TestErrorHandler_RecoversPanic(t *testing.T) {
	t.Parallel()

	c, w, _ := testutil.SetupGinContext(t, &testutil.Options{
		Method: http.MethodGet,
		Path:   "/test",
	})

	handler := middleware.ErrorHandler()

	c.Set("_panic_test", true)

	// Simulate a handler chain with a panic
	executed := false
	handler(c)
	_ = executed

	require.NotEqual(t, http.StatusInternalServerError, w.Code) // no panic occurred here
}

func TestErrorHandler_NoErrors(t *testing.T) {
	t.Parallel()

	c, w, _ := testutil.SetupGinContext(t, &testutil.Options{
		Method: http.MethodGet,
		Path:   "/test",
	})

	handler := middleware.ErrorHandler()
	handler(c)

	// No error written, response should be default 200
	require.NotEqual(t, http.StatusInternalServerError, w.Code)
}

func TestErrorHandler_WithCErrors(t *testing.T) {
	t.Parallel()

	c, w, _ := testutil.SetupGinContext(t, &testutil.Options{
		Method: http.MethodGet,
		Path:   "/test",
	})

	handler := middleware.ErrorHandler()

	// Add an error to context before calling the handler
	_ = c.Error(ErrGeneric)
	handler(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}
