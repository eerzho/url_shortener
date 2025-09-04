package helper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"url_shortener/internal/constant"
	"url_shortener/internal/dto"
	"url_shortener/internal/handler/response"

	"github.com/go-playground/validator/v10"
)

var (
	l *slog.Logger
	v *validator.Validate
)

func Setup(logger *slog.Logger, validate *validator.Validate) {
	l = logger
	v = validate
}

func ParseJSON(request any, body io.Reader) error {
	err := json.NewDecoder(body).Decode(request)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	if v != nil {
		err = v.Struct(request)
		if err != nil {
			return fmt.Errorf("validate: %w", err)
		}
	}
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		if l != nil {
			l.Error("failed to encode response",
				slog.Any("error", err),
				slog.Int("status", status),
				slog.Any("response", response),
			)
		}
	}
}

func OK(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, response.Ok{Data: data})
}

func List(w http.ResponseWriter, list any, pagination *dto.Pagination) {
	WriteJSON(w, http.StatusOK, response.List{Data: list, Pagination: pagination})
}

func Fail(w http.ResponseWriter, err error) {
	status := mapErrToStatus(err)

	level := slog.LevelDebug
	if status >= http.StatusInternalServerError {
		level = slog.LevelError
	}

	if l != nil {
		l.LogAttrs(context.Background(), level, "error occurred",
			slog.Any("error", err),
			slog.Int("status", status),
		)
	}

	WriteJSON(w, status, createFailResponse(err, status))
}

func mapErrToStatus(err error) int {
	var validateErrs validator.ValidationErrors
	if errors.As(err, &validateErrs) {
		return http.StatusBadRequest
	}

	switch {
	case errors.Is(err, constant.ErrAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, constant.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusRequestTimeout
	case errors.Is(err, context.Canceled):
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}

func createFailResponse(err error, status int) *response.Fail {
	failResponse := response.Fail{}
	var validateErrs validator.ValidationErrors
	if errors.As(err, &validateErrs) {
		failResponse.Errors = make([]string, len(validateErrs))
		for i, e := range validateErrs {
			failResponse.Errors[i] = e.Error()
		}
		return &failResponse
	}
	failResponse.Error = http.StatusText(status)
	if status < http.StatusInternalServerError {
		failResponse.Error = unwrapErr(err).Error()
	}
	return &failResponse
}

func unwrapErr(err error) error {
	for errors.Unwrap(err) != nil {
		err = errors.Unwrap(err)
	}
	return err
}
