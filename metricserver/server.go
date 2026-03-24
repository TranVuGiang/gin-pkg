package metricserver

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

const (
	metricsPath = "/metrics"
	statusPath  = "/status"
)

type Config struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type Server struct {
	httpServer *http.Server
}

func New(cfg *Config) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc(statusPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.Handle(metricsPath, promhttp.Handler())

	address := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	return &Server{
		httpServer: &http.Server{
			Addr:         address,
			Handler:      mux,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
	}
}

func (s *Server) Run() {
	log.Info().Str("address", s.httpServer.Addr).Msg("Starting metrics server")

	if err := s.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Panic().Err(err).Msg("Metrics server encountered a fatal error")
	}
}

// Stop shuts down the metrics server using the context deadline provided by the caller (Runner).
func (s *Server) Stop(ctx context.Context) error {
	log.Info().Msg("Initiating graceful shutdown of metrics server")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to gracefully shut down metrics server")

		return err
	}

	log.Info().Msg("Metrics server shutdown complete")

	return nil
}

func (s *Server) Name() string {
	return "metric"
}
