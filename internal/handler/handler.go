package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"url_shortener/internal/service"

	"github.com/eerzho/simpledi"
	"github.com/go-playground/validator/v10"
)

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
		return nil
	}
	return nil
}

func successResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, err error) {
	var response struct {
		Error  string   `json:"error,omitempty"`
		Errors []string `json:"errors,omitempty"`
	}

	var validateErrs validator.ValidationErrors
	if errors.As(err, &validateErrs) {
		for _, e := range validateErrs {
			response.Errors = append(response.Errors, e.Error())
		}
	} else {
		response.Error = err.Error()
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}
