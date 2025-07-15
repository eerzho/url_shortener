package app

import (
	"log/slog"
	"url_shortener/internal/config"
	"url_shortener/internal/handler"
	"url_shortener/internal/handler/middleware"
	postgresrepo "url_shortener/internal/repository/postgres"
	valkeyrepo "url_shortener/internal/repository/valkey"
	"url_shortener/internal/service"
	postgresutils "url_shortener/internal/utils/postgres"
	valkeyutils "url_shortener/internal/utils/valkey"

	"github.com/eerzho/simpledi"
	"github.com/jmoiron/sqlx"
	valkeygo "github.com/valkey-io/valkey-go"
)

type component struct {
	key     string
	needs   []string
	builder func() any
}

func Setup(logger *slog.Logger) {
	simpledi.Register("logger", nil, func() any {
		return logger
	})

	for _, c := range components() {
		simpledi.Register(c.key, c.needs, c.builder)
	}

	simpledi.MustResolve()
}

func Close(logger *slog.Logger) {
	if err := simpledi.MustGetAs[*sqlx.DB]("postgres").Close(); err != nil {
		logger.Error("failed to close postgres", slog.Any("error", err))
	}
	simpledi.MustGetAs[valkeygo.Client]("valkey").Close()
	simpledi.MustGetAs[*service.URL]("url_service").Close()
}

func components() []component {
	return []component{
		{
			"config",
			[]string{"logger"},
			func() any {
				return config.MustNewConfig(
					simpledi.MustGetAs[*slog.Logger]("logger"),
				)
			},
		},
		{
			"postgres",
			[]string{"logger", "config"},
			func() any {
				cfg := simpledi.MustGetAs[*config.Config]("config")
				return postgresutils.MustNewPostgresDB(
					simpledi.MustGetAs[*slog.Logger]("logger"),
					cfg.Postgres.URL,
					cfg.Postgres.MaxOpenConns,
					cfg.Postgres.MaxIdleConns,
					cfg.Postgres.ConnMaxLifetime,
				)
			},
		},
		{
			"valkey",
			[]string{"logger", "config"},
			func() any {
				return valkeyutils.MustNewValkeyClient(
					simpledi.MustGetAs[*slog.Logger]("logger"),
					simpledi.MustGetAs[*config.Config]("config").Valkey.URL,
				)
			},
		},
		// repository
		{
			"url_postgres_repository",
			[]string{"postgres"},
			func() any {
				return postgresrepo.NewURL(
					simpledi.MustGetAs[*sqlx.DB]("postgres"),
				)
			},
		},
		{
			"click_postgres_repository",
			[]string{"postgres"},
			func() any {
				return postgresrepo.NewClick(
					simpledi.MustGetAs[*sqlx.DB]("postgres"),
				)
			},
		},
		{
			"url_valkey_repository",
			[]string{"config", "logger", "valkey", "url_postgres_repository"},
			func() any {
				urlRepo := simpledi.MustGetAs[*postgresrepo.URL]("url_postgres_repository")
				return valkeyrepo.NewURL(
					simpledi.MustGetAs[*config.Config]("config").TTL.URLCache,
					simpledi.MustGetAs[*slog.Logger]("logger"),
					simpledi.MustGetAs[valkeygo.Client]("valkey"),
					urlRepo,
					urlRepo,
				)
			},
		},
		// service
		{
			"url_service",
			[]string{"config", "logger", "url_valkey_repository", "click_postgres_repository"},
			func() any {
				cfg := simpledi.MustGetAs[*config.Config]("config")
				urlRepo := simpledi.MustGetAs[*valkeyrepo.URL]("url_valkey_repository")
				return service.NewURL(
					cfg.WorkerPool.URLCount,
					cfg.WorkerPool.URLBufferSize,
					simpledi.MustGetAs[*slog.Logger]("logger"),
					urlRepo,
					urlRepo,
					simpledi.MustGetAs[*postgresrepo.Click]("click_postgres_repository"),
				)
			},
		},
		{
			"ip_service",
			nil,
			func() any {
				return service.NewIP()
			},
		},
		{
			"click_service",
			[]string{"config", "click_postgres_repository"},
			func() any {
				cfg := simpledi.MustGetAs[*config.Config]("config")
				return service.NewClick(
					cfg.Pagination.MinPage,
					cfg.Pagination.MinSize,
					cfg.Pagination.MaxSize,
					simpledi.MustGetAs[*postgresrepo.Click]("click_postgres_repository"),
				)
			},
		},
		// middleware
		{
			"rate_limiter_middleware",
			[]string{"config", "ip_service"},
			func() any {
				cfg := simpledi.MustGetAs[*config.Config]("config")
				return middleware.NewRateLimiter(
					cfg.RateLimit.RPS,
					cfg.RateLimit.Burst,
					cfg.RateLimit.CacheCapacity,
					simpledi.MustGetAs[*service.IP]("ip_service"),
				)
			},
		},
		{
			"logger_middleware",
			[]string{"logger", "ip_service"},
			func() any {
				return middleware.NewLogger(
					simpledi.MustGetAs[*slog.Logger]("logger"),
					simpledi.MustGetAs[*service.IP]("ip_service"),
				)
			},
		},
		// handler
		{
			"handler",
			[]string{"logger"},
			func() any {
				return handler.New(
					simpledi.MustGetAs[*slog.Logger]("logger"),
				)
			},
		},
		{
			"url_handler",
			[]string{"handler", "url_service", "ip_service"},
			func() any {
				return handler.NewURL(
					simpledi.MustGetAs[*handler.Handler]("handler"),
					simpledi.MustGetAs[*service.URL]("url_service"),
					simpledi.MustGetAs[*service.IP]("ip_service"),
				)
			},
		},
		{
			"click_handler",
			[]string{"handler", "click_service"},
			func() any {
				return handler.NewClick(
					simpledi.MustGetAs[*handler.Handler]("handler"),
					simpledi.MustGetAs[*service.Click]("click_service"),
				)
			},
		},
	}
}
