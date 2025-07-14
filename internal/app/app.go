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

func Setup(logger *slog.Logger) error {
	simpledi.Register("logger", nil, func() any {
		return logger
	})

	setupUtils()
	setupRepository()
	setupService()
	setupMiddleware()
	setupHandler()

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

func setupUtils() {
	simpledi.Register(
		"config",
		[]string{"logger"},
		func() any {
			return config.MustNewConfig(
				simpledi.Get("logger").(*slog.Logger),
			)
		},
	)
	simpledi.Register(
		"postgres",
		[]string{"logger", "config"},
		func() any {
			return utilspostgres.MustNewPostgresDB(
				simpledi.Get("logger").(*slog.Logger),
				simpledi.Get("config").(*config.Config).Postgres.URL,
			)
		},
	)
	simpledi.Register(
		"valkey",
		[]string{"logger", "config"},
		func() any {
			return utilsvalkey.MustNewValkeyClient(
				simpledi.Get("logger").(*slog.Logger),
				simpledi.Get("config").(*config.Config).Valkey.URL,
			)
		},
	)
}

func setupRepository() {
	simpledi.Register(
		"url_postgres_repository",
		[]string{"postgres"},
		func() any {
			return repositorypostgres.NewURL(
				simpledi.Get("postgres").(*sqlx.DB),
			)
		},
	)
	simpledi.Register(
		"click_postgres_repository",
		[]string{"postgres"},
		func() any {
			return repositorypostgres.NewClick(
				simpledi.Get("postgres").(*sqlx.DB),
			)
		},
	)
	simpledi.Register(
		"url_valkey_repository",
		[]string{"logger", "valkey", "url_postgres_repository"},
		func() any {
			return repositoryvalkey.NewURL(
				simpledi.Get("logger").(*slog.Logger),
				simpledi.Get("valkey").(valkeygo.Client),
				simpledi.Get("url_postgres_repository").(*repositorypostgres.URL),
			)
		},
	)
}

func setupService() {
	simpledi.Register(
		"url_service",
		[]string{"logger", "url_valkey_repository", "click_postgres_repository"},
		func() any {
			return service.NewURL(
				simpledi.Get("logger").(*slog.Logger),
				simpledi.Get("url_valkey_repository").(*repositoryvalkey.URL),
				simpledi.Get("click_postgres_repository").(*repositorypostgres.Click),
			)
		},
	)
	simpledi.Register(
		"ip_service",
		nil,
		func() any {
			return service.NewIP()
		},
	)
	simpledi.Register(
		"click_service",
		[]string{"click_postgres_repository"},
		func() any {
			return service.NewClick(
				simpledi.Get("click_postgres_repository").(*repositorypostgres.Click),
			)
		},
	)
}

func setupMiddleware() {
	simpledi.Register(
		"rate_limiter_middleware",
		[]string{"ip_service"},
		func() any {
			return middleware.NewRateLimiter(
				simpledi.Get("ip_service").(*service.IP),
			)
		},
	)
	simpledi.Register(
		"logger_middleware",
		[]string{"logger", "ip_service"},
		func() any {
			return middleware.NewLogger(
				simpledi.Get("logger").(*slog.Logger),
				simpledi.Get("ip_service").(*service.IP),
			)
		},
	)
}

func setupHandler() {
	simpledi.Register(
		"handler",
		[]string{"logger"},
		func() any {
			return handler.New(
				simpledi.Get("logger").(*slog.Logger),
			)
		},
	)
	simpledi.Register(
		"url_handler",
		[]string{"handler", "url_service", "ip_service"},
		func() any {
			return handler.NewURL(
				simpledi.Get("handler").(*handler.Handler),
				simpledi.Get("url_service").(*service.URL),
				simpledi.Get("ip_service").(*service.IP),
			)
		},
	)
	simpledi.Register(
		"click_handler",
		[]string{"handler", "click_service"},
		func() any {
			return handler.NewClick(
				simpledi.Get("handler").(*handler.Handler),
				simpledi.Get("click_service").(*service.Click),
			)
		},
	)
}
