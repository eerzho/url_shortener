package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"url_shortener/internal/handler/response"
	"url_shortener/internal/service"

	"github.com/eerzho/simpledi"
	"github.com/go-playground/validator/v10"
)

// @title           url shortener api
// @version         1.0
// @BasePath        /
func Setup(mux *http.ServeMux, c *simpledi.Container) {
	mux.HandleFunc("POST /urls", urlCreate(c.Get("url_service").(service.Url)))
	mux.HandleFunc("GET /urls/{short_code}", urlShow(c.Get("url_service").(service.Url)))
	mux.HandleFunc("GET /{short_code}", urlRedirect(c.Get("url_service").(service.Url)))
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
