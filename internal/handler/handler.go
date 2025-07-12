package handler

import (
	"context"
	"encoding/json"
	"errors"
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

func (h *Handler) parseJson(request any, body io.Reader) error {
	err := json.NewDecoder(body).Decode(request)
	if err != nil {
		return err
	}
	err = h.v.Struct(request)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) writeJson(w http.ResponseWriter, status int, response any) {
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
	h.writeJson(w, status, response.Ok{Data: data})
}

func (h *Handler) list(w http.ResponseWriter, list any, pagination *dto.Pagination) {
	h.writeJson(w, http.StatusOK, response.List{Data: list, Pagination: pagination})
}

func (h *Handler) fail(w http.ResponseWriter, err error) {
	status := h.mapErrorToStatus(err)
	response := response.Fail{}

	var validateErrs validator.ValidationErrors
	if errors.As(err, &validateErrs) {
		status = http.StatusBadRequest
		response.Errors = make([]string, len(validateErrs))
		for i, e := range validateErrs {
			response.Errors[i] = e.Error()
		}
	} else {
		response.Error = err.Error()
	}

	h.writeJson(w, status, response)
}

func (h *Handler) mapErrorToStatus(err error) int {
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
	urlHandler := simpledi.Get("url_handler").(*Url)
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
