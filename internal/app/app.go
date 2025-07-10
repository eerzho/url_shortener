package app

import (
	"url_shortener/internal/config"
	"url_shortener/internal/handler/middleware"
	"url_shortener/internal/repository/postgres"
	"url_shortener/internal/service"
	"url_shortener/internal/utils"

	"github.com/eerzho/simpledi"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	valkeygo "github.com/valkey-io/valkey-go"
)

func Setup() {
	simpledi.Register("config", nil, func() any {
		return config.NewConfig()
	})
	simpledi.Register("postgres", []string{"config"}, func() any {
		return utils.NewPostgresCLient(
			simpledi.Get("config").(*config.Config).Postgres.Url,
		)
	})
	simpledi.Register("valkey", []string{"config"}, func() any {
		return utils.NewValkeyClient(
			simpledi.Get("config").(*config.Config).Valkey.Url,
		)
	})
	simpledi.Register("url_postgres_repository", []string{"postgres"}, func() any {
		return postgres.NewUrl(
			simpledi.Get("postgres").(*sqlx.DB),
		)
	})
	simpledi.Register("url_service", []string{"url_postgres_repository"}, func() any {
		return service.NewUrl(
			simpledi.Get("url_postgres_repository").(*postgres.Url),
		)
	})
	simpledi.Register("rate_limiter_middleware", nil, func() any {
		return middleware.NewRateLimiter(1_000, 10, 20)
	})
	simpledi.Register("logger_middleware", nil, func() any {
		return middleware.NewLogger()
	})

	err := simpledi.Resolve()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to resolve dependencies")
	}
}

func Close() {
	simpledi.Get("postgres").(*sqlx.DB).Close()
	simpledi.Get("valkey").(valkeygo.Client).Close()
	// simpledi.Get("url_valkey_repository").(*valkey.Url).Close()
}
