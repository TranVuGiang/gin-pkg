package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type LogFieldExtractor func(*gin.Context) map[string]any

var logFieldsPool = sync.Pool{
	New: func() any {
		return make(map[string]any, 10)
	},
}

func RequestLogger(log zerolog.Logger, extraLogFieldExtractor ...LogFieldExtractor) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		fields := logFieldsPool.Get().(map[string]any)
		defer func() {
			clear(fields)
			logFieldsPool.Put(fields)
		}()

		fillLogFields(fields, c, start)

		if id, exists := c.Get(ContextKeyRequestID); exists {
			if idStr, ok := id.(string); ok && idStr != "" {
				fields["request_id"] = idStr
			}
		}

		for _, extractor := range extraLogFieldExtractor {
			for k, v := range extractor(c) {
				fields[k] = v
			}
		}

		logRequest(log, fields, c.Writer.Status())
	}
}

func fillLogFields(fields map[string]any, c *gin.Context, start time.Time) {
	fields["remote_ip"] = c.ClientIP()
	fields["latency"] = time.Since(start).String()
	fields["host"] = c.Request.Host
	fields["request"] = c.Request.Method + " " + c.Request.URL.String()
	fields["request_uri"] = c.Request.RequestURI
	fields["status"] = c.Writer.Status()
	fields["size"] = c.Writer.Size()
	fields["user_agent"] = c.Request.UserAgent()
}

func logRequest(log zerolog.Logger, fields map[string]any, status int) {
	logger := log.With().Fields(fields).Logger()

	switch {
	case status >= http.StatusInternalServerError:
		logger.Error().Msg("Server error")
	case status >= http.StatusBadRequest:
		logger.Warn().Msg("Client error")
	case status >= http.StatusMultipleChoices:
		logger.Info().Msg("Redirection")
	default:
		logger.Info().Msg("Success")
	}
}
