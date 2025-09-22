package helper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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

	err = v.Struct(request)
	if err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, status int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		l.Error("failed to encode response",
			slog.Any("error", err),
			slog.Int("status", status),
			slog.Any("response", response),
		)
	}
}

func Ok(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, response.NewOk(data))
}

func Fail(w http.ResponseWriter, err error) {
	status := mapErrToStatus(err)

	level := slog.LevelDebug
	if status >= http.StatusInternalServerError {
		level = slog.LevelError
	}

	l.LogAttrs(context.Background(), level, "error occurred",
		slog.Any("error", err),
		slog.Int("status", status),
	)

	WriteJSON(w, status, response.NewFail(status, err))
}

func mapErrToStatus(err error) int {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return http.StatusBadRequest
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusRequestTimeout
	case errors.Is(err, context.Canceled):
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}
