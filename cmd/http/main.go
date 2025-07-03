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

	swagger "github.com/swaggo/http-swagger"
)

func main() {
	c := app.Setup()
	defer app.Close(c)

	mux := http.NewServeMux()
	handler.Setup(mux, c)
	mux.Handle("/swagger/", swagger.WrapHandler)

	server := &http.Server{
		Addr:         ":" + c.Get("config").(*config.Config).Http.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("starting http server on http://localhost%s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
