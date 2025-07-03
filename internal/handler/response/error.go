package response

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

type Error struct {
	Error  string   `json:"error,omitempty"`
	Errors []string `json:"errors,omitempty"`
}

func NewError(err error) *Error {
	var validateErrs validator.ValidationErrors
	if errors.As(err, &validateErrs) {
		errs := make([]string, 0, len(validateErrs))
		for _, e := range validateErrs {
			errs = append(errs, e.Error())
		}
		return &Error{
			Errors: errs,
		}
	} else {
		return &Error{
			Error: err.Error(),
		}
	}
}
