package app

import (
	"log/slog"
	"time"
	"url_shortener/internal/config"
	"url_shortener/internal/handler"
	"url_shortener/internal/handler/middleware"
	postgresRepo "url_shortener/internal/repository/postgres"
	valkeyRepo "url_shortener/internal/repository/valkey"
	"url_shortener/internal/service"
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
		{
			Key:  "urlPostgresRepo",
			Deps: []string{"postgres"},
			Ctor: func() any {
				db := simpledi.MustGetAs[*sqlx.DB]("postgres")
				return postgresRepo.NewURL(
					db,
				)
			},
		},
		{
			Key:  "counterValkeyRepo",
			Deps: []string{"valkey"},
			Ctor: func() any {
				client := simpledi.MustGetAs[valkeygo.Client]("valkey")
				return valkeyRepo.NewCounter(
					client,
				)
			},
		},
		{
			Key:  "urlValkeyRepo",
			Deps: []string{"logger", "valkey", "urlPostgresRepo"},
			Ctor: func() any {
				logger := simpledi.MustGetAs[*slog.Logger]("logger")
				client := simpledi.MustGetAs[valkeygo.Client]("valkey")
				urlRepo := simpledi.MustGetAs[*postgresRepo.URL]("urlPostgresRepo")
				return valkeyRepo.NewURL(
					time.Hour*24,
					logger,
					client,
					urlRepo,
				)
			},
		},
		{
			Key:  "urlService",
			Deps: []string{"urlValkeyRepo", "counterValkeyRepo"},
			Ctor: func() any {
				urlRepo := simpledi.MustGetAs[*valkeyRepo.URL]("urlValkeyRepo")
				counterRepo := simpledi.MustGetAs[*valkeyRepo.Counter]("counterValkeyRepo")
				return service.NewURL(
					urlRepo,
					counterRepo,
				)
			},
		},
		{
			Key:  "loggerMiddleware",
			Deps: []string{"logger"},
			Ctor: func() any {
				logger := simpledi.MustGetAs[*slog.Logger]("logger")
				return middleware.NewLogger(
					logger,
				)
			},
		},
		{
			Key:  "urlHandler",
			Deps: []string{"urlService"},
			Ctor: func() any {
				urlService := simpledi.MustGetAs[*service.URL]("urlService")
				return handler.NewURL(
					urlService,
				)
			},
		},
	}
}
