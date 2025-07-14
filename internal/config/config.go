package config

import (
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
)

type (
	Config struct {
		APP      APP
		HTTP     HTTP
		Postgres Postgres
		Valkey   Valkey
	}

	APP struct {
		ENV string `env:"APP_ENV" envDefault:"prod"`
	}

	HTTP struct {
		Port string `env:"HTTP_PORT,required"`
	}

	Postgres struct {
		URL string `env:"POSTGRES_URL,required"`
	}

	Valkey struct {
		URL string `env:"VALKEY_URL,required"`
	}
)

func NewConfig() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func MustNewConfig(logger *slog.Logger) *Config {
	cfg, err := NewConfig()
	if err != nil {
		logger.Error("failed to parse to env", slog.Any("error", err))
		os.Exit(1)
	}
	return cfg
}
