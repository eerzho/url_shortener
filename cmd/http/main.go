package main

import (
	"context"
	"log"
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
	app.MustSetup()
	defer app.Close()

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
		log.Printf("starting http server on port: %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("http server failed: %v\n", err)
			return
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("shutting down server...\n")

	ctx, cancel := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v\n", err)
		return
	}

	log.Printf("http server exited\n")
}
