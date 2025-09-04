package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type (
	Config struct {
		APP        APP
		HTTP       HTTP
		Postgres   Postgres
		Valkey     Valkey
		TTL        TTL
		Pagination Pagination
	}

	APP struct {
		ENV string `env:"APP_ENV" envDefault:"prod"`
	}

	HTTP struct {
		Port           string        `env:"HTTP_PORT"            envDefault:"8080"`
		ReadTimeout    time.Duration `env:"HTTP_READ_TIMEOUT"    envDefault:"10s"`
		WriteTimeout   time.Duration `env:"HTTP_WRITE_TIMEOUT"   envDefault:"10s"`
		IdleTimeout    time.Duration `env:"HTTP_IDLE_TIMEOUT"    envDefault:"60s"`
		RequestTimeout time.Duration `env:"HTTP_REQUEST_TIMEOUT" envDefault:"30s"`
	}

	Postgres struct {
		URL             string        `env:"POSTGRES_URL,required"`
		MaxOpenConns    int           `env:"POSTGRES_MAX_OPEN_CONNS"    envDefault:"25"`
		MaxIdleConns    int           `env:"POSTGRES_MAX_IDLE_CONNS"    envDefault:"5"`
		ConnMaxLifetime time.Duration `env:"POSTGRES_CONN_MAX_LIFETIME" envDefault:"5m"`
	}

	Valkey struct {
		URL string `env:"VALKEY_URL,required"`
	}

	TTL struct {
		URLCache time.Duration `env:"TTL_URL_CACHE" envDefault:"24h"`
	}

	Pagination struct {
		MinPage int `env:"PAGINATION_MIN_PAGE" envDefault:"1"`
		MinSize int `env:"PAGINATION_MIN_SIZE" envDefault:"5"`
		MaxSize int `env:"PAGINATION_MAX_SIZE" envDefault:"200"`
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
		panic(err)
	}
	return cfg
}
