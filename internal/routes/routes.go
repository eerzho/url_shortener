package routes

import (
	"net/http"
	"url_shortener/internal/handler"
	"url_shortener/internal/service"

	"github.com/eerzho/simpledi"
)

func Setup(mux *http.ServeMux, c *simpledi.Container) {
	mux.HandleFunc("POST /urls", handler.Create(c.Get("url_service").(*service.Url)))
	mux.HandleFunc("GET /urls/{short_code}", handler.Show(c.Get("url_service").(*service.Url)))
	mux.HandleFunc("GET /{short_code}", handler.Redirect(c.Get("url_service").(*service.Url)))
}
