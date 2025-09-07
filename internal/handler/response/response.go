package response

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type Ok struct {
	Data any `json:"data"`
}

type Fail struct {
	Error  string   `json:"error,omitempty"`
	Errors []string `json:"errors,omitempty"`
}

func NewOk(data any) *Ok {
	return &Ok{
		Data: data,
	}
}

func NewFail(status int, err error) *Fail {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		responseErrs := make([]string, len(validationErrs))
		for i, validationErr := range validationErrs {
			responseErrs[i] = validationErr.Error()
		}
		return &Fail{Errors: responseErrs}
	}

	responseErr := http.StatusText(status)
	if status < http.StatusInternalServerError {
		responseErr = unwrapErr(err).Error()
	}

	return &Fail{Error: responseErr}
}

func unwrapErr(err error) error {
	for errors.Unwrap(err) != nil {
		err = errors.Unwrap(err)
	}
	return err
}
