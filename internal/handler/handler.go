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

type Url struct {
	Id        int       `db:"id" json:"id"`
	ShortCode string    `db:"short_code" json:"short_code"`
	LongUrl   string    `db:"long_url" json:"long_url"`
	Clicks    int       `db:"clicks" json:"clicks"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func Create(c *simpledi.Container) http.HandlerFunc {
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

		postgres := c.Get("postgres").(*sqlx.DB)

		var url Url
		shortCode := fmt.Sprintf("%x", sha256.Sum256([]byte(request.LongUrl)))[:6]
		err = postgres.Get(
			&url,
			`
				insert into urls (short_code, long_url)
				values ($1, $2)
				returning *
			`,
			shortCode, request.LongUrl,
		)
		if err != nil {
			errorResponse(w, err)
			return
		}

		successResponse(w, http.StatusCreated, &url)
	}
}

func Show(c *simpledi.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("short_code")

		if shortCode == "" {
			errorResponse(w, fmt.Errorf("short_code is required"))
			return
		}

		postgres := c.Get("postgres").(*sqlx.DB)

		var url Url
		err := postgres.Get(
			&url,
			`select * from urls where short_code = $1`,
			shortCode,
		)
		if err != nil {
			errorResponse(w, err)
			return
		}

		successResponse(w, http.StatusOK, &url)
	}
}

func Redirect(c *simpledi.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("short_code")

		if shortCode == "" {
			errorResponse(w, fmt.Errorf("short_code is required"))
			return
		}

		postgres := c.Get("postgres").(*sqlx.DB)

		var url Url
		err := postgres.Get(
			&url,
			`
				update urls set clicks = clicks + 1, updated_at = now()
				where short_code = $1
				returning *
			`,
			shortCode,
		)
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
