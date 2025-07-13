package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"url_shortener/internal/constant"
	"url_shortener/internal/dto"
	"url_shortener/internal/handler/middleware"
	"url_shortener/internal/handler/response"

	"github.com/eerzho/simpledi"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	v *validator.Validate
}

func New() *Handler {
	return &Handler{
		v: validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (h *Handler) parseJSON(request any, body io.Reader) error {
	err := json.NewDecoder(body).Decode(request)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	err = h.v.Struct(request)
	if err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().
			Err(err).
			Int("status", status).
			Any("response", response).
			Msg("failed to encode response")
	}
}

func (h *Handler) ok(w http.ResponseWriter, status int, data any) {
	h.writeJSON(w, status, response.Ok{Data: data})
}

func (h *Handler) list(w http.ResponseWriter, list any, pagination *dto.Pagination) {
	h.writeJSON(w, http.StatusOK, response.List{Data: list, Pagination: pagination})
}

func (h *Handler) fail(w http.ResponseWriter, err error) {
	status := h.mapErrToStatus(err)

	logger := log.Debug()
	if status >= 500 {
		logger = log.Error()
	}
	logger.Err(err).
		Int("status", status).
		Msg("error occurred")

	h.writeJSON(w, status, h.createFailResponse(err, status))
}

func (h *Handler) unwrapErr(err error) error {
	for errors.Unwrap(err) != nil {
		err = errors.Unwrap(err)
	}
	return err
}

func (h *Handler) createFailResponse(err error, status int) *response.Fail {
	response := response.Fail{}
	var validateErrs validator.ValidationErrors
	if errors.As(err, &validateErrs) {
		response.Errors = make([]string, len(validateErrs))
		for i, e := range validateErrs {
			response.Errors[i] = e.Error()
		}
		return &response
	}
	response.Error = http.StatusText(status)
	if status < 500 {
		response.Error = h.unwrapErr(err).Error()
	}
	return &response
}

func (h *Handler) mapErrToStatus(err error) int {
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

func Setup(mux *http.ServeMux) {
	rateLimiterMiddleware := simpledi.Get("rate_limiter_middleware").(*middleware.RateLimiter)
	loggerMiddleware := simpledi.Get("logger_middleware").(*middleware.Logger)
	urlHandler := simpledi.Get("url_handler").(*URL)
	clickHandler := simpledi.Get("click_handler").(*Click)

	mux.Handle("POST /urls", middleware.ChainFunc(
		urlHandler.Create,
		loggerMiddleware.Handle,
		rateLimiterMiddleware.Handle,
	))
	mux.Handle("GET /urls/{short_code}", middleware.ChainFunc(
		urlHandler.Stats,
		loggerMiddleware.Handle,
		rateLimiterMiddleware.Handle,
	))
	mux.Handle("GET /urls/{short_code}/clicks", middleware.ChainFunc(
		clickHandler.List,
		loggerMiddleware.Handle,
		rateLimiterMiddleware.Handle,
	))
	mux.Handle("GET /{short_code}", middleware.ChainFunc(
		urlHandler.Click,
		loggerMiddleware.Handle,
		rateLimiterMiddleware.Handle,
	))
}
