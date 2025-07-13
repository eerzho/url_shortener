package config

import (
	"log"

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

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatalf("failed to parse to env: %v\n", err)
	}
	return cfg
}
