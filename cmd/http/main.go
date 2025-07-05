package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "url_shortener/docs"
	"url_shortener/internal/app"
	"url_shortener/internal/config"
	"url_shortener/internal/handler"

	"github.com/eerzho/simpledi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	env := os.Getenv("APP_ENV")
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "prod" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if env == "prod" || env == "stage" {
		log.Logger = zerolog.New(os.Stdout).With().
			Timestamp().
			Str("service", "url_shortener").
			Str("app_env", env).
			Logger()
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().
			Str("service", "url_shortener").
			Logger()
	}
}

func main() {
	app.Setup()
	defer app.Close()

	mux := http.NewServeMux()
	handler.Setup(mux)

	server := &http.Server{
		Addr:         ":" + simpledi.Get("config").(*config.Config).Http.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("port", server.Addr).Msg("starting http server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("http server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("http server exited")
}
