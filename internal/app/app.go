package app

import (
	"url_shortener/internal/config"
	"url_shortener/internal/handler"
	"url_shortener/internal/handler/middleware"
	"url_shortener/internal/repository/postgres"
	"url_shortener/internal/repository/valkey"
	"url_shortener/internal/service"
	utilspostgres "url_shortener/internal/utils/postgres"
	utilsvalkey "url_shortener/internal/utils/valkey"

	"github.com/eerzho/simpledi"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	valkeygo "github.com/valkey-io/valkey-go"
)

func Setup() {
	setupUtils()
	setupRepository()
	setupService()
	setupMiddleware()
	setupHandler()

	err := simpledi.Resolve()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to resolve dependencies")
	}
}

func setupUtils() {
	simpledi.Register(
		"config",
		nil,
		func() any {
			return config.NewConfig()
		},
	)
	simpledi.Register(
		"postgres",
		[]string{"config"},
		func() any {
			return utilspostgres.NewPostgresDB(
				simpledi.Get("config").(*config.Config).Postgres.URL,
			)
		},
	)
	simpledi.Register(
		"valkey",
		[]string{"config"},
		func() any {
			return utilsvalkey.NewValkeyClient(
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
			return postgres.NewURL(
				simpledi.Get("postgres").(*sqlx.DB),
			)
		},
	)
	simpledi.Register(
		"click_postgres_repository",
		[]string{"postgres"},
		func() any {
			return postgres.NewClick(
				simpledi.Get("postgres").(*sqlx.DB),
			)
		},
	)
	simpledi.Register(
		"url_valkey_repository",
		[]string{"valkey", "url_postgres_repository"},
		func() any {
			return valkey.NewURL(
				simpledi.Get("valkey").(valkeygo.Client),
				simpledi.Get("url_postgres_repository").(*postgres.URL),
			)
		},
	)
}

func setupService() {
	simpledi.Register(
		"url_service",
		[]string{"url_valkey_repository", "click_postgres_repository"},
		func() any {
			return service.NewURL(
				simpledi.Get("url_valkey_repository").(*valkey.URL),
				simpledi.Get("click_postgres_repository").(*postgres.Click),
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
				simpledi.Get("click_postgres_repository").(*postgres.Click),
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
				10,
				20,
				1_000,
			)
		},
	)
	simpledi.Register(
		"logger_middleware",
		[]string{"ip_service"},
		func() any {
			return middleware.NewLogger(
				simpledi.Get("ip_service").(*service.IP),
			)
		},
	)
}

func setupHandler() {
	simpledi.Register(
		"handler",
		nil,
		func() any {
			return handler.New()
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

func Close() {
	if err := simpledi.Get("postgres").(*sqlx.DB).Close(); err != nil {
		log.Error().Err(err).Msg("failed to close postgres")
	}
	simpledi.Get("valkey").(valkeygo.Client).Close()
	simpledi.Get("url_service").(*service.URL).Close()
}
