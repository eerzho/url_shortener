package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func NewLogger(env string) zerolog.Logger {
	var level zerolog.Level
	if env == "prod" {
		level = zerolog.InfoLevel
	} else {
		level = zerolog.DebugLevel
	}

	var logger zerolog.Logger
	if env == "prod" || env == "stage" {
		logger = zerolog.New(os.Stdout).
			Level(level).
			With().
			Timestamp().
			Str("app_env", env).
			Logger()
	} else {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).
			Level(level).
			With().
			Timestamp().
			Str("app_env", env).
			Logger()
	}

	return logger
}
