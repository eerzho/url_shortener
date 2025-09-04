package handler

import (
	"log/slog"
	"net/http"
	"url_shortener/internal/handler/helper"
	"url_shortener/internal/handler/middleware"

	"github.com/eerzho/simpledi"
	"github.com/go-playground/validator/v10"
)

func Setup(mux *http.ServeMux) {
	logger := simpledi.MustGetAs[*slog.Logger]("logger")
	validate := simpledi.MustGetAs[*validator.Validate]("validate")
	urlHandler := simpledi.MustGetAs[*URL]("urlHandler")
	clickHandler := simpledi.MustGetAs[*Click]("clickHandler")
	loggerMiddleware := simpledi.MustGetAs[*middleware.Logger]("loggerMiddleware")

	helper.Setup(logger, validate)

	mux.Handle("POST /urls", middleware.ChainFunc(
		urlHandler.Create,
		loggerMiddleware.Handle,
	))
	mux.Handle("GET /urls/{short_code}", middleware.ChainFunc(
		urlHandler.Stats,
		loggerMiddleware.Handle,
	))
	mux.Handle("GET /urls/{short_code}/clicks", middleware.ChainFunc(
		clickHandler.List,
		loggerMiddleware.Handle,
	))
	mux.Handle("GET /{short_code}", middleware.ChainFunc(
		urlHandler.Click,
		loggerMiddleware.Handle,
	))
}
