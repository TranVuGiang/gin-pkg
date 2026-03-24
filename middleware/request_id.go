package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Skipper func(*gin.Context) bool

func DefaultSkipper(_ *gin.Context) bool {
	return false
}

func RequestID(skipper Skipper) gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipper(c) {
			c.Next()
			return
		}

		rid := strings.TrimSpace(c.GetHeader(HeaderXRequestID))

		if rid == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprint("missing required header: ", HeaderXRequestID),
			})

			return
		}

		if err := uuid.Validate(rid); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("invalid %s: must be a valid UUID", HeaderXRequestID),
			})

			return
		}

		c.Set(ContextKeyRequestID, rid)
		c.Next()
	}
}
