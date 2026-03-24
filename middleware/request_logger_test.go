package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TranVuGiang/gin_pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func BenchmarkRequestLogger(b *testing.B) {
	gin.SetMode(gin.TestMode)

	mw := middleware.RequestLogger(log.Logger)

	b.ResetTimer()

	for range make([]struct{}, b.N) {
		w := httptest.NewRecorder()
		c, engine := gin.CreateTestContext(w)

		engine.Use(mw)
		engine.GET("/", func(ctx *gin.Context) {
			time.Sleep(5 * time.Millisecond)
			ctx.String(http.StatusOK, "OK")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request = req

		engine.ServeHTTP(w, req)
	}
}
