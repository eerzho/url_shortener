package utils

import (
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

func NewPostgresCLient(url string) *sqlx.DB {
	client, err := sqlx.Connect("postgres", url)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}

	client.SetMaxOpenConns(25)
	client.SetMaxIdleConns(5)
	client.SetConnMaxLifetime(5 * time.Minute)

	return client
}
