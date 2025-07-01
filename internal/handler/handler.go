package handler

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/eerzho/simpledi"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

var validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())

func Shorten(container *simpledi.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			LongUrl string `json:"long_url" validate:"required,url"`
		}

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			errorResponse(w, err)
			return
		}

		err = validate.Struct(request)
		if err != nil {
			errorResponse(w, err)
			return
		}

		shortCode := fmt.Sprintf("%x", sha256.Sum256([]byte(request.LongUrl)))[:6]

		var url struct {
			Id        int       `db:"id" json:"id"`
			ShortCode string    `db:"short_code" json:"short_code"`
			LongUrl   string    `db:"long_url" json:"long_url"`
			Clicks    int       `db:"clicks" json:"clicks"`
			CreatedAt time.Time `db:"created_at" json:"created_at"`
			UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
		}

		postgres := container.Get("postgres").(*sqlx.DB)

		err = postgres.Get(
			&url,
			"insert into urls (short_code, long_url) values ($1, $2) returning id, short_code, long_url, clicks, created_at, updated_at",
			shortCode, request.LongUrl,
		)
		if err != nil {
			errorResponse(w, err)
			return
		}

		successResponse(w, http.StatusOK, &url)
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
