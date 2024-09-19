package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"go.xsfx.dev/caddy-log-exporter/internal/config"
)

const metricBaseName = "caddy_log_exporter"

type Server struct {
	addr       string
	logFiles   []string
	httpServer *http.Server
}

func New(c config.Config) *Server {
	s := &Server{}

	s.addr = c.Addr
	s.logFiles = c.LogFiles

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", s.metrics)

	s.httpServer = &http.Server{
		Addr:              c.Addr,
		Handler:           mux,
		ReadHeaderTimeout: time.Second,
	}

	return s
}

func (s *Server) metrics(w http.ResponseWriter, _ *http.Request) {
	metrics.WritePrometheus(w, false)
}

const shutdownTimeout = 10 * time.Second

func (s *Server) http(ctx context.Context, wg *sync.WaitGroup) {
	go func() {
		slog.Info("serving http", "addr", s.addr)

		if err := s.httpServer.ListenAndServe(); !errors.Is(
			err,
			http.ErrServerClosed,
		) {
			slog.Error("http server error", "err", err)

			return
		}

		slog.Info("stopped http server")
	}()

	wg.Add(1)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil { //nolint:contextcheck
			slog.Error("http server shutdown", "err", err)
		}

		wg.Done()
	}()
}

func (s *Server) Serve(ctx context.Context) error {
	if err := s.Tail(parseFunc); err != nil {
		return fmt.Errorf("starting tailing: %w", err)
	}

	wg := &sync.WaitGroup{}

	s.http(ctx, wg)

	wg.Wait()

	return nil
}
