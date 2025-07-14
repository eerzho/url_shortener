package logger

import (
	"log/slog"
	"os"
	"slices"
)

const (
	EnvProd  = "prod"
	EnvStage = "stage"
	EnvDev   = "dev"
)

func NewLogger(env string) *slog.Logger {
	if !slices.Contains([]string{EnvProd, EnvStage, EnvDev}, env) {
		env = EnvProd
	}

	var level slog.Level
	if env == EnvProd {
		level = slog.LevelInfo
	} else {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: env == EnvDev,
	}

	var handler slog.Handler
	if env == EnvProd || env == EnvStage {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler).With(
		slog.String("app_env", env),
	)
}
