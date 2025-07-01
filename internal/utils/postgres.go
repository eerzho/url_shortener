package utils

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgresCLient(url string) *sqlx.DB {
	client, err := sqlx.Connect("postgres", url)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	return client
}
