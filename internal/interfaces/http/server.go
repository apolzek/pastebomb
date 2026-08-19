package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/apolzek/pastebomb/internal/infrastructure/config"
)

// Server owns the listener lifecycle, including a graceful drain so in-flight
// requests are not cut off mid-response on deploy.
type Server struct {
	http     *http.Server
	shutdown config.HTTP
	limiters Limiters
}

// NewServer wraps the engine in an http.Server with explicit timeouts. The
// defaults of http.Server are all zero, which means "no timeout" and leaves
// the process open to slow-client exhaustion.
func NewServer(engine *gin.Engine, limiters Limiters, cfg config.HTTP) *Server {
	return &Server{
		http: &http.Server{
			Addr:              cfg.Address,
			Handler:           engine,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		},
		shutdown: cfg,
		limiters: limiters,
	}
}

// Run serves until ctx is cancelled, then drains and returns.
func (s *Server) Run(ctx context.Context) error {
	errs := make(chan error, 1)

	go func() {
		slog.Info("http server listening", "address", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
			return
		}
		errs <- nil
	}()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
		return s.drain()
	}
}

func (s *Server) drain() error {
	slog.Info("http server shutting down", "timeout", s.shutdown.ShutdownTimeout)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdown.ShutdownTimeout)
	defer cancel()

	err := s.http.Shutdown(shutdownCtx)
	s.limiters.Stop()
	return err
}
