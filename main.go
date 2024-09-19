package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"go.xsfx.dev/caddy-log-exporter/internal/config"
	"go.xsfx.dev/caddy-log-exporter/internal/logger"
	"go.xsfx.dev/caddy-log-exporter/internal/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	logger.Setup()

	cfg, err := config.Parse()
	if err != nil {
		slog.Error("parsing config", "err", err)
		cancel()
		os.Exit(1)
	}

	s := server.New(cfg)
	if err := s.Serve(ctx); err != nil {
		slog.Error("serving", "err", err)
		cancel()
		os.Exit(1)
	}

	cancel()
}
