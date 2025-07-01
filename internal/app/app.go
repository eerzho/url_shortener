package app

import (
	"database/sql"
	"log"
	"url_shortener/internal/config"
	"url_shortener/internal/utils"

	"github.com/eerzho/simpledi"
	"github.com/valkey-io/valkey-go"
)

type App struct {
	Config    *config.Config
	Container *simpledi.Container
}

func NewApp() *App {
	cfg := config.NewConfig()
	container := simpledi.NewContainer()

	container.Register("postgres", nil, func() any {
		return utils.NewPostgresCLient(cfg.Postgres.Url)
	})
	container.Register("valkey", nil, func() any {
		return utils.NewValkeyClient(cfg.Valkey.Url)
	})

	err := container.Resolve()
	if err != nil {
		log.Fatalf("failed to resolve dependencies: %v", err)
	}

	return &App{
		Config:    cfg,
		Container: container,
	}
}

func (a *App) Close() {
	if db, ok := a.Container.Get("db").(*sql.DB); ok {
		db.Close()
	}

	if cache, ok := a.Container.Get("cache").(valkey.Client); ok {
		cache.Close()
	}
}
