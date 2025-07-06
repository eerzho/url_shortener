package response

import (
	"errors"
	"net/http"
	"url_shortener/internal/constant"

	"github.com/go-playground/validator/v10"
)

type Error struct {
	Error      string   `json:"error,omitempty"`
	Errors     []string `json:"errors,omitempty"`
	StatusCode int      `json:"-"`
}

func NewError(err error) *Error {
	var errResponse Error

	errResponse.StatusCode = http.StatusInternalServerError
	switch {
	case errors.Is(err, constant.ErrAlreadyExists):
		errResponse.StatusCode = http.StatusConflict
	case errors.Is(err, constant.ErrNotFound):
		errResponse.StatusCode = http.StatusNotFound
	}

	var validateErrs validator.ValidationErrors
	if errors.As(err, &validateErrs) {
		errResponse.StatusCode = http.StatusBadRequest
		errResponse.Errors = make([]string, 0, len(validateErrs))
		for _, e := range validateErrs {
			errResponse.Errors = append(errResponse.Errors, e.Error())
		}
	} else {
		errResponse.Error = err.Error()
	}

	return &errResponse
}
