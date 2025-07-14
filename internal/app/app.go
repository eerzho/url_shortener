package app

import (
	"log/slog"
	"os"
	"url_shortener/internal/config"
	"url_shortener/internal/handler"
	"url_shortener/internal/handler/middleware"
	repositorypostgres "url_shortener/internal/repository/postgres"
	repositoryvalkey "url_shortener/internal/repository/valkey"
	"url_shortener/internal/service"
	utilspostgres "url_shortener/internal/utils/postgres"
	utilsvalkey "url_shortener/internal/utils/valkey"

	"github.com/eerzho/simpledi"
	"github.com/jmoiron/sqlx"
	valkeygo "github.com/valkey-io/valkey-go"
)

type component struct {
	key     string
	needs   []string
	builder func() any
}

func Setup(logger *slog.Logger) error {
	simpledi.Register("logger", nil, func() any {
		return logger
	})
	for _, c := range components() {
		simpledi.Register(c.key, c.needs, c.builder)
	}
	err := simpledi.Resolve()
	if err != nil {
		return err
	}
	return nil
}

func Close(logger *slog.Logger) {
	if err := simpledi.Get("postgres").(*sqlx.DB).Close(); err != nil {
		logger.Error("failed to close postgres", slog.Any("error", err))
	}
	simpledi.Get("valkey").(valkeygo.Client).Close()
	simpledi.Get("url_service").(*service.URL).Close()
}

func MustSetup(logger *slog.Logger) {
	err := Setup(logger)
	if err != nil {
		logger.Error("failed to setup app", slog.Any("error", err))
		os.Exit(1)
	}
}

func components() []component {
	return []component{
		{
			"config",
			[]string{"logger"},
			func() any {
				return config.MustNewConfig(
					simpledi.Get("logger").(*slog.Logger),
				)
			},
		},
		{
			"postgres",
			[]string{"logger", "config"},
			func() any {
				return utilspostgres.MustNewPostgresDB(
					simpledi.Get("logger").(*slog.Logger),
					simpledi.Get("config").(*config.Config).Postgres.URL,
				)
			},
		},
		{
			"valkey",
			[]string{"logger", "config"},
			func() any {
				return utilsvalkey.MustNewValkeyClient(
					simpledi.Get("logger").(*slog.Logger),
					simpledi.Get("config").(*config.Config).Valkey.URL,
				)
			},
		},
		// repository
		{
			"url_postgres_repository",
			[]string{"postgres"},
			func() any {
				return repositorypostgres.NewURL(
					simpledi.Get("postgres").(*sqlx.DB),
				)
			},
		},
		{
			"click_postgres_repository",
			[]string{"postgres"},
			func() any {
				return repositorypostgres.NewClick(
					simpledi.Get("postgres").(*sqlx.DB),
				)
			},
		},
		{
			"url_valkey_repository",
			[]string{"logger", "valkey", "url_postgres_repository"},
			func() any {
				return repositoryvalkey.NewURL(
					simpledi.Get("logger").(*slog.Logger),
					simpledi.Get("valkey").(valkeygo.Client),
					simpledi.Get("url_postgres_repository").(*repositorypostgres.URL),
					simpledi.Get("url_postgres_repository").(*repositorypostgres.URL),
				)
			},
		},
		// service
		{
			"url_service",
			[]string{"logger", "url_valkey_repository", "click_postgres_repository"},
			func() any {
				return service.NewURL(
					simpledi.Get("logger").(*slog.Logger),
					simpledi.Get("url_valkey_repository").(*repositoryvalkey.URL),
					simpledi.Get("url_valkey_repository").(*repositoryvalkey.URL),
					simpledi.Get("click_postgres_repository").(*repositorypostgres.Click),
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
			[]string{"click_postgres_repository"},
			func() any {
				return service.NewClick(
					simpledi.Get("click_postgres_repository").(*repositorypostgres.Click),
				)
			},
		},
		// middleware
		{
			"rate_limiter_middleware",
			[]string{"ip_service"},
			func() any {
				return middleware.NewRateLimiter(
					simpledi.Get("ip_service").(*service.IP),
				)
			},
		},
		{
			"logger_middleware",
			[]string{"logger", "ip_service"},
			func() any {
				return middleware.NewLogger(
					simpledi.Get("logger").(*slog.Logger),
					simpledi.Get("ip_service").(*service.IP),
				)
			},
		},
		// handler
		{
			"handler",
			[]string{"logger"},
			func() any {
				return handler.New(
					simpledi.Get("logger").(*slog.Logger),
				)
			},
		},
		{
			"url_handler",
			[]string{"handler", "url_service", "ip_service"},
			func() any {
				return handler.NewURL(
					simpledi.Get("handler").(*handler.Handler),
					simpledi.Get("url_service").(*service.URL),
					simpledi.Get("ip_service").(*service.IP),
				)
			},
		},
		{
			"click_handler",
			[]string{"handler", "click_service"},
			func() any {
				return handler.NewClick(
					simpledi.Get("handler").(*handler.Handler),
					simpledi.Get("click_service").(*service.Click),
				)
			},
		},
	}
}
