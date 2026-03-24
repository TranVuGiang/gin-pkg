package httpserver

import (
	"github.com/TranVuGiang/gin_pkg/middleware"
	"github.com/TranVuGiang/gin_pkg/notifylog"
	"github.com/gin-gonic/gin"
)

// HandlerFunc is the typed delegate called by ExecuteStandardized.
// The handler receives a structured logger, the Gin context, and a bound+validated request.
// It returns a HandlerResponse (with optional Status and Pagination) or an AppError.
type HandlerFunc[REQ any, RES any] func(log notifylog.NotifyLog, c *gin.Context, request *REQ) (*HandlerResponse[RES], *AppError)

// ExecuteStandardized calls delegate, wraps the result in APIResponse, and returns it.
// Use this when you need pagination, a request-scoped logger, or the standard response envelope.
// For simpler handlers, use Wrapper directly.
func ExecuteStandardized[REQ any, RES any](c *gin.Context, request *REQ, handlerName string, delegate HandlerFunc[REQ, RES]) (any, *AppError) {
	log := notifylog.New(handlerName, notifylog.JSON)

	log.Info().Str("handler", handlerName).Msg("handler invoked")

	internalResponse, delegateError := delegate(log, c, request)
	if delegateError != nil {
		log.Error().
			Int("status", delegateError.Code).
			Any("cause", delegateError.Internal).
			Msgf("Request failed with HTTP error: %s", delegateError.Message)

		return nil, delegateError
	}

	requestID, _ := middleware.GetRequestID(c)
	if requestID == "" {
		requestID = "N/A"
	}

	finalPayload := &APIResponse[RES]{
		RequestID:  requestID,
		Data:       internalResponse.Data,
		Pagination: internalResponse.Pagination,
	}

	log.Info().Str("handler", handlerName).Msg("handler completed")

	return finalPayload, nil
}
