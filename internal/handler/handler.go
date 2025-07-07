package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"
	"url_shortener/internal/constant"
	"url_shortener/internal/handler/middleware"
	"url_shortener/internal/handler/response"
	"url_shortener/internal/metrics"
	"url_shortener/internal/service"

	"github.com/eerzho/simpledi"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	swagger "github.com/swaggo/http-swagger"
)

// @title           URL Shortener API
// @version         1.0
// @description     A simple URL shortener service
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
// @license.name   Apache 2.0
// @license.url    http://www.apache.org/licenses/LICENSE-2.0.html
// @host           localhost:8080
// @BasePath       /
// @schemes        http https
func Setup(mux *http.ServeMux) {
	urlService := simpledi.Get("url_service").(*service.Url)

	// System endpoints first (most specific)
	mux.Handle("GET /swagger/", middleware.ChainFunc(swagger.WrapHandler, middleware.Logging))
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.Handle("GET /health", middleware.ChainFunc(healthCheck(urlService), middleware.Logging))

	// API endpoints (specific paths)
	mux.Handle("POST /urls", middleware.ChainFunc(urlCreate(urlService), middleware.Logging, middleware.RateLimit))
	mux.Handle("GET /urls/{short_code}", middleware.ChainFunc(urlShow(urlService), middleware.Logging, middleware.RateLimit))
	mux.Handle("GET /urls/{short_code}/stats", middleware.ChainFunc(urlStats(urlService), middleware.Logging, middleware.RateLimit))

	// Redirect endpoint (catch-all, must be last)
	mux.Handle("GET /{short_code}", middleware.ChainFunc(urlRedirect(urlService), middleware.Logging, middleware.RateLimit))
}

var validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())

// Custom validator interface for request validation
type Validator interface {
	Validate() error
}

// decodeAndValidate decodes JSON request body and validates it
func decodeAndValidate(request interface{}, body io.Reader) error {
	// Decode JSON
	if err := json.NewDecoder(body).Decode(request); err != nil {
		return &ValidationError{
			Field:   "json",
			Message: "Invalid JSON format",
			Err:     err,
		}
	}

	// Validate using validator package
	if err := validate.Struct(request); err != nil {
		return err
	}

	// Custom validation if the struct implements Validator interface
	if validator, ok := request.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return &ValidationError{
				Field:   "custom",
				Message: err.Error(),
				Err:     err,
			}
		}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (v *ValidationError) Error() string {
	return v.Message
}

func (v *ValidationError) Unwrap() error {
	return v.Err
}

// successResponse sends a successful JSON response
func successResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("failed to encode response")
		// If encoding fails, we can't send a proper error response
		// since headers are already written
	}
}

// errorResponse sends an error JSON response
func errorResponse(w http.ResponseWriter, r *http.Request, err error) {
	requestID := getRequestID(r.Context())
	errResponse := response.NewErrorWithContext(err, requestID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errResponse.StatusCode)

	if encodeErr := json.NewEncoder(w).Encode(errResponse); encodeErr != nil {
		log.Error().Err(encodeErr).Msg("failed to encode error response")
	}
}

// getRequestID extracts request ID from context
func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(constant.RequestIdKey).(string); ok {
		return requestID
	}
	return ""
}

// recordHTTPMetrics records HTTP request metrics
func recordHTTPMetrics(method, endpoint string, statusCode int, duration time.Duration) {
	metrics.RecordHTTPRequest(method, endpoint, strconv.Itoa(statusCode), duration)
}

// healthCheck godoc
// @Summary      Health check endpoint
// @Description  Check if the service is healthy
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      503  {object}  response.Error
// @Router       /health [get]
func healthCheck(urlService UrlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			recordHTTPMetrics(r.Method, "/health", http.StatusOK, time.Since(start))
		}()

		logger := log.With().
			Str("op", "handler.healthCheck").
			Str("request_id", getRequestID(r.Context())).
			Logger()

		logger.Debug().Msg("performing health check")

		// Check service health
		if err := urlService.HealthCheck(r.Context()); err != nil {
			logger.Error().Err(err).Msg("service health check failed")
			errorResponse(w, r, err)
			return
		}

		healthStatus := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
			"uptime":    time.Since(start).String(),
		}

		logger.Debug().Msg("health check passed")
		successResponse(w, http.StatusOK, healthStatus)
	}
}

// urlCreate godoc
// @Summary      Create short URL
// @Description  Create a short URL from a long URL
// @Tags         url
// @Accept       json
// @Produce      json
// @Param        input  body      request.CreateUrl  true  "URL creation request"
// @Success      201    {object}  response.Url
// @Failure      400    {object}  response.Error
// @Failure      409    {object}  response.Error
// @Failure      500    {object}  response.Error
// @Router       /urls [post]
func urlCreate(urlService UrlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		statusCode := http.StatusCreated
		defer func() {
			recordHTTPMetrics(r.Method, "/urls", statusCode, time.Since(start))
		}()

		logger := log.With().
			Str("op", "handler.urlCreate").
			Str("request_id", getRequestID(r.Context())).
			Logger()

		logger.Debug().Msg("creating short URL")

		var request struct {
			LongUrl string `json:"long_url" validate:"required,url,min=10,max=2048"`
		}

		if err := decodeAndValidate(&request, r.Body); err != nil {
			logger.Debug().Err(err).Msg("request validation failed")
			statusCode = http.StatusBadRequest
			errorResponse(w, r, err)
			return
		}

		logger.Debug().Str("long_url", request.LongUrl).Msg("validated request")

		url, err := urlService.Create(r.Context(), request.LongUrl)
		if err != nil {
			logger.Error().Err(err).Str("long_url", request.LongUrl).Msg("failed to create short URL")
			statusCode = http.StatusInternalServerError
			errorResponse(w, r, err)
			return
		}

		logger.Info().
			Int("id", url.Id).
			Str("short_code", url.ShortCode).
			Str("long_url", url.LongUrl).
			Msg("short URL created successfully")

		successResponse(w, http.StatusCreated, response.NewUrl(url))
	}
}

// urlShow godoc
// @Summary      Get URL information
// @Description  Get information about a short URL
// @Tags         url
// @Produce      json
// @Param        short_code  path      string  true  "Short code"
// @Success      200         {object}  response.Url
// @Failure      400         {object}  response.Error
// @Failure      404         {object}  response.Error
// @Failure      500         {object}  response.Error
// @Router       /urls/{short_code} [get]
func urlShow(urlService UrlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		statusCode := http.StatusOK
		defer func() {
			recordHTTPMetrics(r.Method, "/urls/{short_code}", statusCode, time.Since(start))
		}()

		shortCode := r.PathValue("short_code")
		logger := log.With().
			Str("op", "handler.urlShow").
			Str("request_id", getRequestID(r.Context())).
			Str("short_code", shortCode).
			Logger()

		logger.Debug().Msg("getting URL information")

		if shortCode == "" {
			logger.Debug().Msg("short code is empty")
			statusCode = http.StatusBadRequest
			errorResponse(w, r, &ValidationError{
				Field:   "short_code",
				Message: "short code is required",
			})
			return
		}

		url, err := urlService.GetByShortCode(r.Context(), shortCode)
		if err != nil {
			logger.Debug().Err(err).Msg("failed to get URL")
			if err == constant.ErrNotFound {
				statusCode = http.StatusNotFound
			} else {
				statusCode = http.StatusInternalServerError
			}
			errorResponse(w, r, err)
			return
		}

		logger.Debug().
			Int("id", url.Id).
			Str("long_url", url.LongUrl).
			Int("clicks", url.Clicks).
			Msg("URL information retrieved successfully")

		successResponse(w, http.StatusOK, response.NewUrl(url))
	}
}

// urlStats godoc
// @Summary      Get URL statistics
// @Description  Get statistics for a short URL
// @Tags         url
// @Produce      json
// @Param        short_code  path      string  true  "Short code"
// @Success      200         {object}  response.UrlStats
// @Failure      400         {object}  response.Error
// @Failure      404         {object}  response.Error
// @Failure      500         {object}  response.Error
// @Router       /urls/{short_code}/stats [get]
func urlStats(urlService UrlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		statusCode := http.StatusOK
		defer func() {
			recordHTTPMetrics(r.Method, "/urls/{short_code}/stats", statusCode, time.Since(start))
		}()

		shortCode := r.PathValue("short_code")
		logger := log.With().
			Str("op", "handler.urlStats").
			Str("request_id", getRequestID(r.Context())).
			Str("short_code", shortCode).
			Logger()

		logger.Debug().Msg("getting URL statistics")

		if shortCode == "" {
			logger.Debug().Msg("short code is empty")
			statusCode = http.StatusBadRequest
			errorResponse(w, r, &ValidationError{
				Field:   "short_code",
				Message: "short code is required",
			})
			return
		}

		url, err := urlService.GetByShortCode(r.Context(), shortCode)
		if err != nil {
			logger.Debug().Err(err).Msg("failed to get URL")
			if err == constant.ErrNotFound {
				statusCode = http.StatusNotFound
			} else {
				statusCode = http.StatusInternalServerError
			}
			errorResponse(w, r, err)
			return
		}

		stats := response.NewUrlStats(url)
		logger.Debug().
			Int("id", url.Id).
			Int("clicks", url.Clicks).
			Msg("URL statistics retrieved successfully")

		successResponse(w, http.StatusOK, stats)
	}
}

// urlRedirect godoc
// @Summary      Redirect to original URL
// @Description  Redirect to the original URL and increment click count
// @Tags         url
// @Param        short_code  path  string  true  "Short code"
// @Success      302         "Found"
// @Failure      400         {object}  response.Error
// @Failure      404         {object}  response.Error
// @Failure      500         {object}  response.Error
// @Router       /{short_code} [get]
func urlRedirect(urlService UrlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		statusCode := http.StatusFound
		defer func() {
			recordHTTPMetrics(r.Method, "/{short_code}", statusCode, time.Since(start))
		}()

		shortCode := r.PathValue("short_code")
		logger := log.With().
			Str("op", "handler.urlRedirect").
			Str("request_id", getRequestID(r.Context())).
			Str("short_code", shortCode).
			Str("user_agent", r.UserAgent()).
			Str("referer", r.Referer()).
			Logger()

		logger.Debug().Msg("redirecting to original URL")

		if shortCode == "" {
			logger.Debug().Msg("short code is empty")
			statusCode = http.StatusBadRequest
			errorResponse(w, r, &ValidationError{
				Field:   "short_code",
				Message: "short code is required",
			})
			return
		}

		// Skip certain paths that might conflict with redirect
		if shortCode == "health" || shortCode == "metrics" || shortCode == "swagger" {
			logger.Debug().Msg("skipping redirect for system path")
			statusCode = http.StatusNotFound
			errorResponse(w, r, constant.ErrNotFound)
			return
		}

		url, err := urlService.GetByShortCodeAndIncrementClicks(r.Context(), shortCode)
		if err != nil {
			logger.Debug().Err(err).Msg("failed to get URL for redirect")
			if err == constant.ErrNotFound {
				statusCode = http.StatusNotFound
			} else {
				statusCode = http.StatusInternalServerError
			}
			errorResponse(w, r, err)
			return
		}

		logger.Info().
			Int("id", url.Id).
			Str("long_url", url.LongUrl).
			Int("clicks", url.Clicks).
			Msg("redirecting to original URL")

		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Perform redirect
		http.Redirect(w, r, url.LongUrl, http.StatusFound)
	}
}
