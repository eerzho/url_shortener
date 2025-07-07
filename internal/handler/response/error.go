package response

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"url_shortener/internal/constant"

	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// Error represents an error response
type Error struct {
	Error      string            `json:"error,omitempty"`
	Errors     []string          `json:"errors,omitempty"`
	Details    map[string]string `json:"details,omitempty"`
	Code       string            `json:"code,omitempty"`
	StatusCode int               `json:"-"`
	RequestID  string            `json:"request_id,omitempty"`
}

// ErrorCode represents predefined error codes
type ErrorCode string

const (
	CodeValidationError    ErrorCode = "VALIDATION_ERROR"
	CodeNotFound           ErrorCode = "NOT_FOUND"
	CodeAlreadyExists      ErrorCode = "ALREADY_EXISTS"
	CodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	CodeInternalError      ErrorCode = "INTERNAL_ERROR"
	CodeTimeoutError       ErrorCode = "TIMEOUT_ERROR"
	CodeRateLimitError     ErrorCode = "RATE_LIMIT_ERROR"
	CodeBadRequest         ErrorCode = "BAD_REQUEST"
	CodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	CodeForbidden          ErrorCode = "FORBIDDEN"
	CodeConflict           ErrorCode = "CONFLICT"
	CodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// NewError creates a new error response with appropriate status code and details
func NewError(err error) *Error {
	return NewErrorWithContext(err, "")
}

// NewErrorWithContext creates a new error response with request context
func NewErrorWithContext(err error, requestID string) *Error {
	if err == nil {
		return &Error{
			Error:      constant.InternalServerErrorMessage,
			Code:       string(CodeInternalError),
			StatusCode: http.StatusInternalServerError,
			RequestID:  requestID,
		}
	}

	errResponse := &Error{
		StatusCode: http.StatusInternalServerError,
		Code:       string(CodeInternalError),
		RequestID:  requestID,
		Details:    make(map[string]string),
	}

	// Handle validation errors
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		errResponse.StatusCode = http.StatusBadRequest
		errResponse.Code = string(CodeValidationError)
		errResponse.Error = constant.BadRequestMessage
		errResponse.Errors = make([]string, 0, len(validationErrs))

		for _, e := range validationErrs {
			errResponse.Errors = append(errResponse.Errors, formatValidationError(e))
		}

		return errResponse
	}

	// Handle application-specific errors
	switch {
	case errors.Is(err, constant.ErrNotFound):
		errResponse.StatusCode = http.StatusNotFound
		errResponse.Code = string(CodeNotFound)
		errResponse.Error = constant.NotFoundMessage

	case errors.Is(err, constant.ErrAlreadyExists):
		errResponse.StatusCode = http.StatusConflict
		errResponse.Code = string(CodeConflict)
		errResponse.Error = constant.ConflictMessage

	case errors.Is(err, context.DeadlineExceeded):
		errResponse.StatusCode = http.StatusGatewayTimeout
		errResponse.Code = string(CodeTimeoutError)
		errResponse.Error = "Request timeout"
		errResponse.Details["timeout"] = "The request took too long to complete"

	case errors.Is(err, context.Canceled):
		errResponse.StatusCode = http.StatusRequestTimeout
		errResponse.Code = string(CodeTimeoutError)
		errResponse.Error = "Request cancelled"
		errResponse.Details["cancelled"] = "The request was cancelled"

	case errors.Is(err, sql.ErrNoRows):
		errResponse.StatusCode = http.StatusNotFound
		errResponse.Code = string(CodeNotFound)
		errResponse.Error = constant.NotFoundMessage

	default:
		// Handle PostgreSQL errors
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			errResponse = handlePostgreSQLError(pqErr, requestID)
		} else {
			// Check if it's a wrapped error with additional context
			errResponse = handleGenericError(err, requestID)
		}
	}

	// Log error details for debugging
	logError(err, errResponse)

	return errResponse
}

// handlePostgreSQLError handles PostgreSQL-specific errors
func handlePostgreSQLError(pqErr *pq.Error, requestID string) *Error {
	errResponse := &Error{
		StatusCode: http.StatusInternalServerError,
		Code:       string(CodeDatabaseError),
		RequestID:  requestID,
		Details:    make(map[string]string),
	}

	switch pqErr.Code {
	case "23505": // unique_violation
		errResponse.StatusCode = http.StatusConflict
		errResponse.Code = string(CodeAlreadyExists)
		errResponse.Error = constant.ConflictMessage
		errResponse.Details["constraint"] = pqErr.Constraint

	case "23503": // foreign_key_violation
		errResponse.StatusCode = http.StatusBadRequest
		errResponse.Code = string(CodeValidationError)
		errResponse.Error = "Referenced record does not exist"
		errResponse.Details["constraint"] = pqErr.Constraint

	case "23514": // check_violation
		errResponse.StatusCode = http.StatusBadRequest
		errResponse.Code = string(CodeValidationError)
		errResponse.Error = "Data violates check constraint"
		errResponse.Details["constraint"] = pqErr.Constraint

	case "23502": // not_null_violation
		errResponse.StatusCode = http.StatusBadRequest
		errResponse.Code = string(CodeValidationError)
		errResponse.Error = "Required field is missing"
		errResponse.Details["column"] = pqErr.Column

	case "42P01": // undefined_table
		errResponse.StatusCode = http.StatusInternalServerError
		errResponse.Code = string(CodeDatabaseError)
		errResponse.Error = "Database schema error"

	case "42703": // undefined_column
		errResponse.StatusCode = http.StatusInternalServerError
		errResponse.Code = string(CodeDatabaseError)
		errResponse.Error = "Database schema error"

	case "08000", "08003", "08006": // connection errors
		errResponse.StatusCode = http.StatusServiceUnavailable
		errResponse.Code = string(CodeServiceUnavailable)
		errResponse.Error = "Database temporarily unavailable"

	default:
		errResponse.StatusCode = http.StatusInternalServerError
		errResponse.Code = string(CodeDatabaseError)
		errResponse.Error = "Database operation failed"
		errResponse.Details["pg_code"] = string(pqErr.Code)
	}

	return errResponse
}

// handleGenericError handles generic errors with context
func handleGenericError(err error, requestID string) *Error {
	errResponse := &Error{
		StatusCode: http.StatusInternalServerError,
		Code:       string(CodeInternalError),
		RequestID:  requestID,
		Details:    make(map[string]string),
	}

	errStr := err.Error()

	// Check for specific error patterns
	switch {
	case strings.Contains(errStr, "validation"):
		errResponse.StatusCode = http.StatusBadRequest
		errResponse.Code = string(CodeValidationError)
		errResponse.Error = constant.BadRequestMessage

	case strings.Contains(errStr, "not found"):
		errResponse.StatusCode = http.StatusNotFound
		errResponse.Code = string(CodeNotFound)
		errResponse.Error = constant.NotFoundMessage

	case strings.Contains(errStr, "already exists"):
		errResponse.StatusCode = http.StatusConflict
		errResponse.Code = string(CodeAlreadyExists)
		errResponse.Error = constant.ConflictMessage

	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline"):
		errResponse.StatusCode = http.StatusGatewayTimeout
		errResponse.Code = string(CodeTimeoutError)
		errResponse.Error = "Request timeout"

	case strings.Contains(errStr, "unauthorized"):
		errResponse.StatusCode = http.StatusUnauthorized
		errResponse.Code = string(CodeUnauthorized)
		errResponse.Error = "Unauthorized"

	case strings.Contains(errStr, "forbidden"):
		errResponse.StatusCode = http.StatusForbidden
		errResponse.Code = string(CodeForbidden)
		errResponse.Error = "Forbidden"

	case strings.Contains(errStr, "rate limit"):
		errResponse.StatusCode = http.StatusTooManyRequests
		errResponse.Code = string(CodeRateLimitError)
		errResponse.Error = constant.RateLimitExceededMessage

	default:
		errResponse.Error = constant.InternalServerErrorMessage
	}

	return errResponse
}

// formatValidationError formats validator.FieldError into a user-friendly message
func formatValidationError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "numeric":
		return fmt.Sprintf("%s must contain only numbers", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// logError logs error details for debugging
func logError(err error, errResponse *Error) {
	logger := log.With().
		Str("error_code", errResponse.Code).
		Int("status_code", errResponse.StatusCode).
		Str("request_id", errResponse.RequestID).
		Logger()

	if errResponse.StatusCode >= 500 {
		logger.Error().Err(err).Msg("internal server error")
	} else if errResponse.StatusCode >= 400 {
		logger.Warn().Err(err).Msg("client error")
	} else {
		logger.Debug().Err(err).Msg("error response")
	}
}

// NewValidationError creates a validation error response
func NewValidationError(field, message string) *Error {
	return &Error{
		StatusCode: http.StatusBadRequest,
		Code:       string(CodeValidationError),
		Error:      constant.BadRequestMessage,
		Errors:     []string{fmt.Sprintf("%s: %s", field, message)},
	}
}

// NewNotFoundError creates a not found error response
func NewNotFoundError(resource string) *Error {
	return &Error{
		StatusCode: http.StatusNotFound,
		Code:       string(CodeNotFound),
		Error:      constant.NotFoundMessage,
		Details:    map[string]string{"resource": resource},
	}
}

// NewConflictError creates a conflict error response
func NewConflictError(resource string) *Error {
	return &Error{
		StatusCode: http.StatusConflict,
		Code:       string(CodeConflict),
		Error:      constant.ConflictMessage,
		Details:    map[string]string{"resource": resource},
	}
}

// NewRateLimitError creates a rate limit error response
func NewRateLimitError() *Error {
	return &Error{
		StatusCode: http.StatusTooManyRequests,
		Code:       string(CodeRateLimitError),
		Error:      constant.RateLimitExceededMessage,
		Details:    map[string]string{"retry_after": "60"},
	}
}

// NewTimeoutError creates a timeout error response
func NewTimeoutError() *Error {
	return &Error{
		StatusCode: http.StatusGatewayTimeout,
		Code:       string(CodeTimeoutError),
		Error:      "Request timeout",
		Details:    map[string]string{"timeout": "The request took too long to complete"},
	}
}

// NewInternalError creates an internal server error response
func NewInternalError() *Error {
	return &Error{
		StatusCode: http.StatusInternalServerError,
		Code:       string(CodeInternalError),
		Error:      constant.InternalServerErrorMessage,
	}
}
