package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"url_shortener/internal/handler/middleware"
	"url_shortener/internal/handler/response"
	"url_shortener/internal/service"

	"github.com/eerzho/simpledi"
	"github.com/go-playground/validator/v10"
	swagger "github.com/swaggo/http-swagger"
)

// @title           url shortener api
// @version         1.0
// @BasePath        /
func Setup(mux *http.ServeMux) {
	urlService := simpledi.Get("url_service").(*service.Url)
	rateLimitMiddleware := simpledi.Get("rate_limiter_middleware").(*middleware.RateLimiter)
	loggerMiddleware := simpledi.Get("logger_middleware").(*middleware.Logger)

	mux.Handle("/swagger/", swagger.WrapHandler)

	mux.Handle("POST /urls", middleware.ChainFunc(
		urlCreate(urlService),
		loggerMiddleware.Handle,
		rateLimitMiddleware.Handle,
	))
	mux.Handle("GET /urls/{short_code}", middleware.ChainFunc(
		urlShow(urlService),
		loggerMiddleware.Handle,
		rateLimitMiddleware.Handle,
	))
	mux.Handle("GET /{short_code}", middleware.ChainFunc(
		urlRedirect(urlService),
		loggerMiddleware.Handle,
		rateLimitMiddleware.Handle,
	))
}

var validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())

func decodeAndValidate(request any, body io.Reader) error {
	err := json.NewDecoder(body).Decode(request)
	if err != nil {
		return err
	}
	err = validate.Struct(request)
	if err != nil {
		return err
	}
	return nil
}

func successResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, err error) {
	w.Header().Set("content-type", "application/json")
	errResponse := response.NewError(err)
	w.WriteHeader(errResponse.StatusCode)
	json.NewEncoder(w).Encode(errResponse)
}
