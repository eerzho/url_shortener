package app

import (
	"url_shortener/internal/config"
	"url_shortener/internal/handler"
	"url_shortener/internal/handler/middleware"
	"url_shortener/internal/repository/postgres"
	"url_shortener/internal/repository/valkey"
	"url_shortener/internal/service"
	"url_shortener/internal/utils"

	"github.com/eerzho/simpledi"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	valkeygo "github.com/valkey-io/valkey-go"
)

func Setup() {
	// utils
	simpledi.Register("config", nil, func() any {
		return config.NewConfig()
	})
	simpledi.Register("postgres", []string{"config"}, func() any {
		return utils.NewPostgresDb(
			simpledi.Get("config").(*config.Config).Postgres.Url,
		)
	})
	simpledi.Register("valkey", []string{"config"}, func() any {
		return utils.NewValkeyClient(
			simpledi.Get("config").(*config.Config).Valkey.Url,
		)
	})

	// repository
	simpledi.Register("url_postgres_repository", []string{"postgres"}, func() any {
		return postgres.NewUrl(
			simpledi.Get("postgres").(*sqlx.DB),
		)
	})
	simpledi.Register("click_postgres_repository", []string{"postgres"}, func() any {
		return postgres.NewClick(
			simpledi.Get("postgres").(*sqlx.DB),
		)
	})
	simpledi.Register("url_valkey_repository", []string{"valkey", "url_postgres_repository"}, func() any {
		return valkey.New(
			simpledi.Get("valkey").(valkeygo.Client),
			simpledi.Get("url_postgres_repository").(*postgres.Url),
		)
	})

	// service
	simpledi.Register("url_service", []string{"url_valkey_repository", "click_postgres_repository"}, func() any {
		return service.NewUrl(
			simpledi.Get("url_valkey_repository").(*valkey.Url),
			simpledi.Get("click_postgres_repository").(*postgres.Click),
		)
	})
	simpledi.Register("ip_service", nil, func() any {
		return service.NewIp()
	})
	simpledi.Register("click_service", []string{"click_postgres_repository"}, func() any {
		return service.NewClick(
			simpledi.Get("click_postgres_repository").(*postgres.Click),
		)
	})

	// middleware
	simpledi.Register("rate_limiter_middleware", []string{"ip_service"}, func() any {
		return middleware.NewRateLimiter(
			simpledi.Get("ip_service").(*service.Ip),
			10,
			20,
			1_000,
		)
	})
	simpledi.Register("logger_middleware", []string{"ip_service"}, func() any {
		return middleware.NewLogger(
			simpledi.Get("ip_service").(*service.Ip),
		)
	})

	// handler
	simpledi.Register("handler", nil, func() any {
		return handler.New()
	})
	simpledi.Register("url_handler", []string{"handler", "url_service", "ip_service"}, func() any {
		return handler.NewUrl(
			simpledi.Get("handler").(*handler.Handler),
			simpledi.Get("url_service").(*service.Url),
			simpledi.Get("ip_service").(*service.Ip),
		)
	})
	simpledi.Register("click_handler", []string{"handler", "click_service"}, func() any {
		return handler.NewClick(
			simpledi.Get("handler").(*handler.Handler),
			simpledi.Get("click_service").(*service.Click),
		)
	})

	err := simpledi.Resolve()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to resolve dependencies")
	}
}

func Close() {
	simpledi.Get("postgres").(*sqlx.DB).Close()
	simpledi.Get("valkey").(valkeygo.Client).Close()
	simpledi.Get("url_service").(*service.Url).Close()
}
