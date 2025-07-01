package routes

import (
	"net/http"
	"url_shortener/internal/handler"

	"github.com/eerzho/simpledi"
)

func Setup(mux *http.ServeMux, container *simpledi.Container) {
	mux.HandleFunc("POST /shorten", handler.Shorten(container))

}
