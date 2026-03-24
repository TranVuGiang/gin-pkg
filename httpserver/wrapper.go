package httpserver

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/TranVuGiang/gin_pkg/middleware"
	"github.com/TranVuGiang/gin_pkg/notifylog"
	"github.com/TranVuGiang/gin_pkg/validator"
	"github.com/gin-gonic/gin"
	playgroundvalidator "github.com/go-playground/validator/v10"
)

// Wrapper registers a typed handler with Gin. It binds and validates the request,
// then calls wrapped. On success it respects HandlerResponse.Status for the HTTP status code.
// For full request-scoped logging and pagination support, use ExecuteStandardized instead.
func Wrapper[TREQ any](wrapped func(*gin.Context, *TREQ) (any, *AppError)) gin.HandlerFunc {
	// Capture handler name once at registration time — not per-request.
	handlerName := runtime.FuncForPC(reflect.ValueOf(wrapped).Pointer()).Name()

	return func(c *gin.Context) {
		start := time.Now()
		requestURI := c.Request.RequestURI
		requestID := c.GetHeader(middleware.HeaderXRequestID)
		log := notifylog.New("wrapper", notifylog.JSON).With(map[string]string{
			"request_id": requestID,
		})

		logRequestStart(&log, requestURI, handlerName)

		req, appErr := bindAndValidate[TREQ](c, &log, requestURI)
		if appErr != nil {
			c.AbortWithStatusJSON(appErr.Code, ErrorResponse{Message: appErr.Message})
			return
		}

		c.Set(middleware.ContextKeyBody, req)

		res, appErr := wrapped(c, req)
		if appErr != nil {
			c.AbortWithStatusJSON(appErr.Code, ErrorResponse{Message: appErr.Message})
			return
		}

		status := c.Writer.Status()
		logRequestEnd(&log, status, time.Since(start))
		c.JSON(status, res)
	}
}

func logRequestStart(log *notifylog.NotifyLog, path, handler string) {
	log.Info().
		Str("path", path).
		Str("handler", handler).
		Msg("request started - processing incoming request")
}

func logRequestEnd(log *notifylog.NotifyLog, status int, latency time.Duration) {
	log.Info().
		Dur("latency", latency).
		Int("status", status).
		Msg("request completed - response sent to client")
}

func logError(log *notifylog.NotifyLog, err error, path string, req any, msg string) {
	log.Error().
		Err(err).
		Any("path", path).
		Any("request_object", req).
		Msg(msg)
}

func bindAndValidate[TREQ any](c *gin.Context, log *notifylog.NotifyLog, path string) (*TREQ, *AppError) {
	var req TREQ

	if err := c.ShouldBind(&req); err != nil {
		logError(log, err, path, req, "failed to bind request body to the expected structure")

		return nil, &AppError{
			Code:     http.StatusBadRequest,
			Message:  safeBindError(err),
			Internal: err,
		}
	}

	v := validator.DefaultRestValidator()
	if err := v.Validate(&req); err != nil {
		logError(log, err, path, req, "request validation failed")

		return nil, &AppError{
			Code:     http.StatusBadRequest,
			Message:  safeValidationError(err),
			Internal: err,
		}
	}

	return &req, nil
}

// safeBindError returns a client-safe message for binding failures.
func safeBindError(err error) string {
	// Avoid leaking Go internals from JSON parsing errors.
	return fmt.Sprintf("request body is malformed or missing required fields: %s", err.Error())
}

// safeValidationError converts go-playground/validator errors into human-readable messages
// without leaking internal struct field names or package paths.
func safeValidationError(err error) string {
	var ve playgroundvalidator.ValidationErrors
	if !errors.As(err, &ve) {
		return "request validation failed"
	}

	msgs := make([]string, 0, len(ve))

	for _, fe := range ve {
		msgs = append(msgs, fmt.Sprintf("field '%s' failed validation: %s", fe.Field(), fe.Tag()))
	}

	return strings.Join(msgs, "; ")
}
