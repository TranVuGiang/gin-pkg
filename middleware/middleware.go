package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyTenantID      string = "tenantID"
	ContextKeyRateLimitReqs string = "rateLimitRequests"
	ContextKeyRateLimitWin  string = "rateLimitWindow"
	ContextKeyAPIKey        string = "apiKey"
	ContextKeyRequestID     string = "requestID"
	ContextKeyBody          string = "body"
	ContextKeyToken         string = "token"
	ContextKeyClaims        string = "claims"
)

const (
	HeaderXAPIKey         = "X-Api-Key" //nolint:gosec
	HeaderXRequestID      = "X-Request-ID"
	HeaderXGaianSignature = "X-Gaian-Signature"
	HeaderXGaianTimestamp = "X-Gaian-Timestamp"
)

const (
	HeaderRateLimitLimit     = "X-Ratelimit-Limit"
	HeaderRateLimitRemaining = "X-Ratelimit-Remaining"
	HeaderRateLimitReset     = "X-Ratelimit-Reset"
)

var (
	ErrClaimsNotFound            = errors.New("jwt claims not found in context")
	ErrClaimsTypeAssertionFailed = errors.New("claims type assertion failed")
)

func GetTenantID(c *gin.Context) string {
	if val, ok := c.Get(ContextKeyTenantID); ok {
		if s, ok := val.(string); ok {
			return s
		}
	}

	return ""
}

func GetRateLimitRequests(c *gin.Context) int {
	if val, ok := c.Get(ContextKeyRateLimitReqs); ok {
		if n, ok := val.(int); ok {
			return n
		}
	}

	return 0
}

func GetRateLimitWindow(c *gin.Context) int {
	if val, ok := c.Get(ContextKeyRateLimitWin); ok {
		if n, ok := val.(int); ok {
			return n
		}
	}

	return 0
}

func GetAPIKey(c *gin.Context) string {
	if val, ok := c.Get(ContextKeyAPIKey); ok {
		if s, ok := val.(string); ok {
			return s
		}
	}

	return ""
}

// GetRequestID returns the request ID from context and whether it was present.
// Returns ("", false) when the RequestID middleware was not registered or not yet run.
func GetRequestID(c *gin.Context) (string, bool) {
	val, ok := c.Get(ContextKeyRequestID)
	if !ok {
		return "", false
	}

	s, ok := val.(string)

	return s, ok
}

func GetExtendedClaimsFromContext(c *gin.Context) (*ExtendedClaims, error) {
	val, exists := c.Get(ContextKeyClaims)
	if !exists || val == nil {
		return nil, ErrClaimsNotFound
	}

	claims, ok := val.(*ExtendedClaims)
	if !ok {
		return nil, ErrClaimsTypeAssertionFailed
	}

	return claims, nil
}
