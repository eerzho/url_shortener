package app

import (
	"database/sql"
	"url_shortener/internal/config"
	"url_shortener/internal/repository"
	"url_shortener/internal/repository/postgres"
	"url_shortener/internal/service"
	"url_shortener/internal/utils"

	"github.com/eerzho/simpledi"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/valkey-io/valkey-go"
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
	simpledi.Register("url_repository", []string{"postgres"}, func() any {
		return postgres.NewUrl(
			simpledi.Get("postgres").(*sqlx.DB),
		)
	})
	simpledi.Register("url_service", []string{"url_repository"}, func() any {
		return service.NewUrl(
			simpledi.Get("url_repository").(repository.Url),
		)
	})

	err := simpledi.Resolve()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to resolve dependencies")
	}
}

func Close() {
	simpledi.Get("postgres").(*sql.DB).Close()
	simpledi.Get("valkey").(valkey.Client).Close()
}
