package middleware

import (
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/TranVuGiang/gin_pkg/notifylog"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const bearerPrefix = "Bearer "

type ExtendedClaims struct {
	TenantID   string `json:"tenant_id"`
	CustomerID string `json:"customer_id"`
	jwt.RegisteredClaims
}

func (c *ExtendedClaims) GetTenantID() string {
	return c.TenantID
}

func (c *ExtendedClaims) GetCustomerID() string {
	return c.CustomerID
}

type JwksAuth struct {
	log     notifylog.NotifyLog
	keyfunc keyfunc.Keyfunc
}

func NewJwksAuth(
	log notifylog.NotifyLog,
	keyfunc keyfunc.Keyfunc,
) *JwksAuth {
	return &JwksAuth{
		log:     log,
		keyfunc: keyfunc,
	}
}

func (j *JwksAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(HeaderXRequestID)

		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			j.log.Error().
				Str("request_id", requestID).
				Msg("JWT verification failed: missing or malformed Authorization header")

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Authorization header is required",
			})

			return
		}

		tokenString := authHeader[len(bearerPrefix):]

		claims := &ExtendedClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, j.keyfunc.Keyfunc)
		if err != nil || !token.Valid {
			j.log.Error().
				Err(err).
				Str("request_id", requestID).
				Msg("JWT verification failed")

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid token",
			})

			return
		}

		c.Set(ContextKeyClaims, claims)

		j.log.Debug().
			Str("request_id", requestID).
			Msg("JWT token verified successfully")

		c.Next()
	}
}
