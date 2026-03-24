package testutil

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type Options struct {
	Method string
	Path   string
	Body   []byte
}

func SetupGinContext(
	t *testing.T,
	opts *Options,
) (*gin.Context, *httptest.ResponseRecorder, *http.Request) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(opts.Method, opts.Path, bytes.NewBuffer(opts.Body))
	req.Header.Set("Content-Type", "application/json")

	c.Request = req

	return c, w, req
}
