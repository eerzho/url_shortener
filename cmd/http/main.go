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

// @title    url shortener api
// @version  1.0
// @BasePath /
func main() {
	app.MustSetup()
	defer app.Close()

	mux := http.NewServeMux()
	mux.Handle("/swagger/", swagger.WrapHandler)

	handler.Setup(mux)

	server := &http.Server{
		Addr:         ":" + simpledi.Get("config").(*config.Config).HTTP.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v\n", err)
		return
	}

	log.Printf("http server exited\n")
}
