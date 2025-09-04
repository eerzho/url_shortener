package app

import (
	"log/slog"
	"url_shortener/internal/config"
	postgresUtils "url_shortener/internal/utils/postgres"
	validateUtils "url_shortener/internal/utils/validate"
	valkeyUtils "url_shortener/internal/utils/valkey"

	"github.com/eerzho/simpledi"
	"github.com/jmoiron/sqlx"
	valkeygo "github.com/valkey-io/valkey-go"
)

func Setup(logger *slog.Logger) {
	simpledi.MustRegister(simpledi.Def{
		Key: "logger",
		Ctor: func() any {
			return logger
		},
	})

	for _, def := range defs() {
		simpledi.MustRegister(def)
	}

	simpledi.MustResolve()
}

func Reset(logger *slog.Logger) {
	if err := simpledi.Reset(); err != nil {
		logger.Error("failed to reset container", slog.Any("error", err))
	}
}

func defs() []simpledi.Def {
	return []simpledi.Def{
		{
			Key: "config",
			Ctor: func() any {
				return config.MustNewConfig()
			},
		},
		{
			Key:  "postgres",
			Deps: []string{"config"},
			Ctor: func() any {
				cfg := simpledi.MustGetAs[*config.Config]("config")
				return postgresUtils.MustNewPostgresDB(
					cfg.Postgres.URL,
					cfg.Postgres.MaxOpenConns,
					cfg.Postgres.MaxIdleConns,
					cfg.Postgres.ConnMaxLifetime,
				)
			},
			Dtor: func() error {
				db := simpledi.MustGetAs[*sqlx.DB]("postgres")
				return db.Close()
			},
		},
		{
			Key:  "valkey",
			Deps: []string{"config"},
			Ctor: func() any {
				cfg := simpledi.MustGetAs[*config.Config]("config")
				return valkeyUtils.MustNewValkeyClient(
					cfg.Valkey.URL,
				)
			},
			Dtor: func() error {
				client := simpledi.MustGetAs[valkeygo.Client]("valkey")
				client.Close()
				return nil
			},
		},
		{
			Key: "validate",
			Ctor: func() any {
				return validateUtils.NewValidate()
			},
		},
	}
}
