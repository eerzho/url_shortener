package config

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type (
	Config struct {
		Http     Http
		Postgres Postgres
		Valkey   Valkey
	}

	Http struct {
		Port string `env:"HTTP_PORT"`
	}

	Postgres struct {
		Url string `env:"POSTGRES_URL"`
	}

	Valkey struct {
		Url string `env:"VALKEY_URL"`
	}
)

func NewConfig() *Config {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		log.Fatalf("failed to parse to env: %v", err)
	}
	return &cfg
}
