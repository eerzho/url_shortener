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
	"url_shortener/internal/constant"
	"url_shortener/internal/handler"
	"url_shortener/internal/metrics"

	"github.com/eerzho/simpledi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = constant.EnvDev
	}

	zerolog.TimeFieldFormat = time.RFC3339

	// Set log level based on environment
	switch env {
	case constant.EnvProd:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case constant.EnvStage:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Configure logger format based on environment
	if env == constant.EnvProd || env == constant.EnvStage {
		log.Logger = zerolog.New(os.Stdout).With().
			Timestamp().
			Str("app_env", env).
			Str("service", "url_shortener").
			Str("version", "1.0.0").
			Logger()
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().
			Str("app_env", env).
			Str("service", "url_shortener").
			Logger()
	}

	log.Info().Str("env", env).Msg("logger initialized")
}

func main() {
	log.Info().Msg("starting url shortener service")

	// Initialize application dependencies
	if err := initializeApp(); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize application")
	}
	defer cleanupApp()

	// Create HTTP server
	server := createServer()

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Info().Str("addr", server.Addr).Msg("starting HTTP server")
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for interrupt signal or server error
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	case sig := <-shutdown:
		log.Info().Str("signal", sig.String()).Msg("shutdown signal received")

		// Graceful shutdown
		if err := gracefulShutdown(server); err != nil {
			log.Error().Err(err).Msg("graceful shutdown failed")
		}
	}

	log.Info().Msg("url shortener service stopped")
}

func initializeApp() error {
	log.Debug().Msg("initializing application dependencies")

	// Setup dependency injection
	app.Setup()

	// Initialize metrics
	log.Debug().Msg("metrics initialized")

	return nil
}

func cleanupApp() {
	log.Info().Msg("cleaning up application resources")
	app.Close()
	log.Info().Msg("application cleanup completed")
}

func createServer() *http.Server {
	config := simpledi.Get("config").(*config.Config)

	// Create HTTP multiplexer
	mux := http.NewServeMux()

	// Setup routes
	handler.Setup(mux)

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":" + config.Http.Port,
		Handler:      mux,
		ReadTimeout:  constant.DefaultReadTimeout,
		WriteTimeout: constant.DefaultWriteTimeout,
		IdleTimeout:  constant.DefaultIdleTimeout,
		ErrorLog:     nil, // Use zerolog instead of standard log
	}

	// Update connection metrics
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// This would typically get actual connection count from server
			// For now, we'll just update a placeholder value
			metrics.SetActiveConnections(0)
		}
	}()

	return server
}

func gracefulShutdown(server *http.Server) error {
	log.Info().Msg("initiating graceful shutdown")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), constant.DefaultShutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server shutdown failed")
		return err
	}

	log.Info().Msg("HTTP server shutdown completed")
	return nil
}
