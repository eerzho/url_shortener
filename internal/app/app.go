package app

import (
	"database/sql"
	"log"
	"url_shortener/internal/config"
	"url_shortener/internal/utils"

	"github.com/eerzho/simpledi"
	"github.com/valkey-io/valkey-go"
)

func Setup() *simpledi.Container {
	c := simpledi.NewContainer()

	c.Register("config", nil, func() any {
		return config.NewConfig()
	})
	c.Register("postgres", []string{"config"}, func() any {
		return utils.NewPostgresCLient(
			c.Get("config").(*config.Config).Postgres.Url,
		)
	})
	c.Register("valkey", []string{"config"}, func() any {
		return utils.NewValkeyClient(
			c.Get("config").(*config.Config).Valkey.Url,
		)
	})

	err := c.Resolve()
	if err != nil {
		log.Fatalf("failed to resolve dependencies: %v", err)
	}

	return c
}

func Close(c *simpledi.Container) {
	if db, ok := c.Get("postgres").(*sql.DB); ok {
		db.Close()
	}

	if cache, ok := c.Get("valkey").(valkey.Client); ok {
		cache.Close()
	}
}
