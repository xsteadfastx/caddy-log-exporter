package logger

import (
	"log/slog"
	"os"

	"github.com/phsym/console-slog"
)

func Setup() {
	logger := slog.New(
		console.NewHandler(os.Stderr, &console.HandlerOptions{Level: slog.LevelDebug}),
	)

	slog.SetDefault(logger)
}
