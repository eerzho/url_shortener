package logger

import (
	"log/slog"
	"os"
)

func NewLogger(env string) *slog.Logger {
	var level slog.Level
	if env == "prod" {
		level = slog.LevelInfo
	} else {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if env == "prod" || env == "stage" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler).With(
		slog.String("app_env", env),
	)
}
