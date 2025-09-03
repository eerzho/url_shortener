package app

import (
	"log/slog"
	"url_shortener/internal/config"
	"url_shortener/internal/handler"
	"url_shortener/internal/handler/middleware"
	postgresRepo "url_shortener/internal/repository/postgres"
	valkeyRepo "url_shortener/internal/repository/valkey"
	"url_shortener/internal/service"
	postgresUtils "url_shortener/internal/utils/postgres"
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

func Close(logger *slog.Logger) {
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
			Key:  "clickPostgresRepo",
			Deps: []string{"postgres"},
			Ctor: func() any {
				db := simpledi.MustGetAs[*sqlx.DB]("postgres")
				return postgresRepo.NewClick(
					db,
				)
			},
		},
		{
			Key:  "urlValkeyRepo",
			Deps: []string{"config", "logger", "valkey", "urlPostgresRepo"},
			Ctor: func() any {
				cfg := simpledi.MustGetAs[*config.Config]("config")
				logger := simpledi.MustGetAs[*slog.Logger]("logger")
				client := simpledi.MustGetAs[valkeygo.Client]("valkey")
				urlRepo := simpledi.MustGetAs[*postgresRepo.URL]("urlPostgresRepo")
				return valkeyRepo.NewURL(
					cfg.TTL.URLCache,
					logger,
					client,
					urlRepo,
					urlRepo,
				)
			},
		},
		{
			Key:  "urlService",
			Deps: []string{"config", "logger", "urlValkeyRepo", "clickPostgresRepo"},
			Ctor: func() any {
				cfg := simpledi.MustGetAs[*config.Config]("config")
				logger := simpledi.MustGetAs[*slog.Logger]("logger")
				urlRepo := simpledi.MustGetAs[*valkeyRepo.URL]("urlValkeyRepo")
				clickRepo := simpledi.MustGetAs[*postgresRepo.Click]("clickPostgresRepo")
				return service.NewURL(
					cfg.WorkerPool.URLCount,
					cfg.WorkerPool.URLBufferSize,
					logger,
					urlRepo,
					urlRepo,
					clickRepo,
				)
			},
			Dtor: func() error {
				urlService := simpledi.MustGetAs[*service.URL]("urlService")
				urlService.Close()
				return nil
			},
		},
		{
			Key: "ipService",
			Ctor: func() any {
				return service.NewIP()
			},
		},
		{
			Key:  "clickService",
			Deps: []string{"config", "clickPostgresRepo"},
			Ctor: func() any {
				cfg := simpledi.MustGetAs[*config.Config]("config")
				clickRepo := simpledi.MustGetAs[*postgresRepo.Click]("clickPostgresRepo")
				return service.NewClick(
					cfg.Pagination.MinPage,
					cfg.Pagination.MinSize,
					cfg.Pagination.MaxSize,
					clickRepo,
				)
			},
		},
		{
			Key:  "rateLimiterMiddleware",
			Deps: []string{"config", "ipService"},
			Ctor: func() any {
				cfg := simpledi.MustGetAs[*config.Config]("config")
				ipService := simpledi.MustGetAs[*service.IP]("ipService")
				return middleware.NewRateLimiter(
					cfg.RateLimit.RPS,
					cfg.RateLimit.Burst,
					cfg.RateLimit.CacheCapacity,
					ipService,
				)
			},
		},
		{
			Key:  "loggerMiddleware",
			Deps: []string{"logger", "ipService"},
			Ctor: func() any {
				logger := simpledi.MustGetAs[*slog.Logger]("logger")
				ipService := simpledi.MustGetAs[*service.IP]("ipService")
				return middleware.NewLogger(
					logger,
					ipService,
				)
			},
		},
		{
			Key:  "handler",
			Deps: []string{"logger"},
			Ctor: func() any {
				logger := simpledi.MustGetAs[*slog.Logger]("logger")
				return handler.New(
					logger,
				)
			},
		},
		{
			Key:  "urlHandler",
			Deps: []string{"handler", "urlService", "ipService"},
			Ctor: func() any {
				baseHandler := simpledi.MustGetAs[*handler.Handler]("handler")
				urlService := simpledi.MustGetAs[*service.URL]("urlService")
				ipService := simpledi.MustGetAs[*service.IP]("ipService")
				return handler.NewURL(
					baseHandler,
					urlService,
					ipService,
				)
			},
		},
		{
			Key:  "clickHandler",
			Deps: []string{"handler", "clickService"},
			Ctor: func() any {
				baseHandler := simpledi.MustGetAs[*handler.Handler]("handler")
				clickService := simpledi.MustGetAs[*service.Click]("clickService")
				return handler.NewClick(
					baseHandler,
					clickService,
				)
			},
		},
	}
}
