package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog/log"
)

type (
	Config struct {
		HTTP     HTTP
		Postgres Postgres
		Valkey   Valkey
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

func NewConfig() *Config {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse to env")
	}
	return &cfg
}
