package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/TranVuGiang/gin-pkg/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const (
	defaultReadTimeout  = 30 * time.Second
	defaultWriteTimeout = 30 * time.Second
	defaultGracePeriod  = 10 * time.Second
	defaultBodyLimit    = 4 << 20 // 4MB
	maxBodyLimit        = 10 << 30 // 10GB sanity cap
)

type Config struct {
	Host         string
	Port         int
	EnableCors   bool
	CorsOrigins  []string
	BodyLimit    string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	GracePeriod  time.Duration
}

type Server struct {
	address     string
	gracePeriod time.Duration
	httpServer  *http.Server
	Engine      *gin.Engine
	Root        *gin.RouterGroup
}

func New(cfg *Config) *Server {
	applyConfigDefaults(cfg)

	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(middleware.ErrorHandler())
	engine.Use(bodyLimitMiddleware(cfg.BodyLimit))
	engine.Use(middleware.Metrics())
	engine.Use(middleware.RequestLogger(log.Logger, RestLogFieldsExtractor))

	if cfg.EnableCors {
		engine.Use(buildCorsMiddleware(cfg.CorsOrigins))
	}

	root := engine.Group("")
	address := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	srv := &http.Server{
		Addr:         address,
		Handler:      engine,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &Server{
		gracePeriod: cfg.GracePeriod,
		address:     address,
		httpServer:  srv,
		Engine:      engine,
		Root:        root,
	}
}

func (s *Server) Run() {
	log.Info().Str("address", s.address).Msg("Starting HTTP server")

	if err := s.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Panic().Err(err).Msg("HTTP server encountered a fatal error")
	}
}

// Stop shuts down the server using the context deadline provided by the caller (Runner).
func (s *Server) Stop(ctx context.Context) error {
	log.Info().Msg("Initiating graceful shutdown of HTTP server")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to gracefully shut down HTTP server")

		return err
	}

	return nil
}

func (s *Server) Name() string {
	return "http"
}

func RestLogFieldsExtractor(c *gin.Context) map[string]any {
	if req, exists := c.Get(middleware.ContextKeyBody); exists && req != nil {
		var requestPayload string

		if b, err := json.Marshal(req); err != nil {
			requestPayload = fmt.Sprintf("failed to parse request object as string: %+v", err)
		} else {
			requestPayload = string(b)
		}

		return map[string]any{"request_payload": requestPayload}
	}

	return nil
}

func RequestIDSkipper(skip bool) middleware.Skipper {
	return func(_ *gin.Context) bool {
		return skip
	}
}

func applyConfigDefaults(cfg *Config) {
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = defaultReadTimeout
	}

	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = defaultWriteTimeout
	}

	if cfg.GracePeriod == 0 {
		cfg.GracePeriod = defaultGracePeriod
	}
}

func buildCorsMiddleware(origins []string) gin.HandlerFunc {
	if len(origins) == 0 {
		return cors.Default()
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	})
}

func bodyLimitMiddleware(limit string) gin.HandlerFunc {
	maxBytes, err := parseBodyLimit(limit)
	if err != nil {
		log.Warn().Err(err).Str("value", limit).Msg("Invalid BodyLimit value, using default 4MB")

		maxBytes = defaultBodyLimit
	}

	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

func parseBodyLimit(s string) (int64, error) {
	if s == "" {
		return defaultBodyLimit, nil
	}

	s = strings.TrimSpace(s)
	upper := strings.ToUpper(s)

	var (
		n      int64
		err    error
		suffix string
		mult   int64
	)

	switch {
	case strings.HasSuffix(upper, "G"):
		suffix, mult = "G", 1024*1024*1024
	case strings.HasSuffix(upper, "M"):
		suffix, mult = "M", 1024*1024
	case strings.HasSuffix(upper, "K"):
		suffix, mult = "K", 1024
	default:
		n, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parseBodyLimit: cannot parse %q: %w", s, err)
		}
	}

	if suffix != "" {
		n, err = strconv.ParseInt(s[:len(s)-1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parseBodyLimit: cannot parse %q: %w", s, err)
		}

		n *= mult
	}

	if n <= 0 {
		return 0, fmt.Errorf("parseBodyLimit: value must be positive, got %d", n)
	}

	if n > maxBodyLimit {
		log.Warn().Int64("bytes", n).Msg("BodyLimit exceeds 10GB sanity cap, capping at 10GB")

		n = maxBodyLimit
	}

	return n, nil
}
