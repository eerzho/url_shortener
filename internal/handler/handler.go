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
	urlService := simpledi.Get("url_service").(service.Url)

	mux.Handle("/swagger/", middleware.ChainFunc(swagger.WrapHandler, middleware.Logging))

	mux.Handle("POST /urls", middleware.ChainFunc(urlCreate(urlService), middleware.Logging))
	mux.Handle("GET /urls/{short_code}", middleware.ChainFunc(urlShow(urlService), middleware.Logging))
	mux.Handle("GET /{short_code}", middleware.ChainFunc(urlRedirect(urlService), middleware.Logging))
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
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response.NewError(err))
}
