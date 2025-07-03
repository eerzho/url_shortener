package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"url_shortener/internal/service"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())

func Create(urlService *service.Url) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			LongUrl string `json:"long_url" validate:"required,url"`
		}

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			errorResponse(w, err)
			return
		}

		err = validate.Struct(&request)
		if err != nil {
			errorResponse(w, err)
			return
		}

		url, err := urlService.Create(r.Context(), request.LongUrl)
		if err != nil {
			errorResponse(w, err)
			return
		}

		successResponse(w, http.StatusCreated, url)
	}
}

func Show(urlService *service.Url) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, err := urlService.GetByShortCode(r.Context(), r.PathValue("short_code"))
		if err != nil {
			errorResponse(w, err)
			return
		}

		successResponse(w, http.StatusOK, url)
	}
}

func Redirect(urlService *service.Url) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, err := urlService.GetByShortCodeAndIncrementClicks(r.Context(), r.PathValue("short_code"))
		if err != nil {
			errorResponse(w, err)
			return
		}

		http.Redirect(w, r, url.LongUrl, http.StatusFound)
	}
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
