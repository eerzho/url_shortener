package routes

import (
	"net/http"
	"url_shortener/internal/handler"

	"github.com/eerzho/simpledi"
)

func Setup(mux *http.ServeMux, c *simpledi.Container) {
	mux.HandleFunc("POST /urls", handler.Create(c))
	mux.HandleFunc("GET /urls/{short_code}", handler.Show(c))
	mux.HandleFunc("GET /{short_code}", handler.Redirect(c))
}
