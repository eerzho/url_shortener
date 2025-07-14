package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "url_shortener/docs"
	"url_shortener/internal/app"
	"url_shortener/internal/config"
	"url_shortener/internal/handler"
	utilslogger "url_shortener/internal/utils/logger"

	"github.com/eerzho/simpledi"
	swagger "github.com/swaggo/http-swagger"
)

const (
	DefaultReadTimeout    = 10 * time.Second
	DefaultWriteTimeout   = 10 * time.Second
	DefaultIdleTimeout    = 60 * time.Second
	DefaultRequestTimeout = 30 * time.Second
)

// main godoc
//
//	@Title		url shortener api
//	@Version	1.0
//	@BasePath	/
func main() {
	logger := utilslogger.NewLogger(os.Getenv("APP_ENV"))

	app.MustSetup(logger)
	defer app.Close(logger)

	mux := http.NewServeMux()
	mux.Handle("/swagger/", swagger.WrapHandler)

	handler.Setup(mux)

	server := &http.Server{
		Addr:         ":" + simpledi.Get("config").(*config.Config).HTTP.Port,
		Handler:      mux,
		ReadTimeout:  DefaultReadTimeout,
		WriteTimeout: DefaultWriteTimeout,
		IdleTimeout:  DefaultIdleTimeout,
	}

	go func() {
		logger.Info("starting http server", slog.String("port", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", slog.Any("error", err))
			return
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", slog.Any("error", err))
		return
	}

	logger.Info("http server exited")
}
