package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog/log"
)

type (
	Config struct {
		Http     Http
		Postgres Postgres
		Valkey   Valkey
	}

	Http struct {
		Port string `env:"HTTP_PORT,required"`
	}

	Postgres struct {
		Url string `env:"POSTGRES_URL,required"`
	}

	Valkey struct {
		Url string `env:"VALKEY_URL,required"`
	}
)

func NewConfig() *Config {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse to env")
	}
	return &cfg
}
