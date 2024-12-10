package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markonick/gigs-challenge/internal/logger"
)

// ValidationErrorMessages maps validation tags to error messages
var ValidationErrorMessages = map[string]string{
	"required":       "This field is required",
	"eventIDFormat":  "Event ID must start with 'evt_'",
	"validEventType": "Invalid event type",
	"validProject":   "Invalid project identifier",
	"validEventData": "Invalid event data structure",
	"ltefield":       "Created time cannot be in the future",
}

// GetValidationMessage returns the appropriate error message for a validation tag
func GetValidationMessage(tag string) string {
	if msg, exists := ValidationErrorMessages[tag]; exists {
		return msg
	}
	return "Validation failed"
}

// Error types
type (
	ValidationError struct {
		Code   string `json:"code"`
		Detail string `json:"detail"`
	}

	AuthError struct {
		Code   string `json:"code"`
		Detail string `json:"detail"`
	}

	ForbiddenError struct {
		Code   string `json:"code"`
		Detail string `json:"detail"`
	}

	NotFoundError struct {
		Code   string `json:"code"`
		Detail string `json:"detail"`
	}

	ConflictError struct {
		Code   string `json:"code"`
		Detail string `json:"detail"`
	}

	PayloadTooLargeError struct {
		Code   string `json:"code"`
		Detail string `json:"detail"`
	}

	RateLimitError struct {
		Code   string `json:"code"`
		Detail string `json:"detail"`
	}

	InternalError struct {
		Message string `json:"message"`
	}
)

// Error interface implementations
func (e *ValidationError) Error() string      { return e.Detail }
func (e *AuthError) Error() string            { return e.Detail }
func (e *ForbiddenError) Error() string       { return e.Detail }
func (e *NotFoundError) Error() string        { return e.Detail }
func (e *ConflictError) Error() string        { return e.Detail }
func (e *PayloadTooLargeError) Error() string { return e.Detail }
func (e *RateLimitError) Error() string       { return e.Detail }
func (e *InternalError) Error() string        { return e.Message }

// RespondWithError handles different error types and sends appropriate HTTP responses
func RespondWithError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *ValidationError:
		logger.Log.Warn().Err(err).Msg("Validation error")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": e})

	case *AuthError:
		logger.Log.Warn().Err(err).Msg("Authentication error")
		c.JSON(http.StatusUnauthorized, gin.H{"error": e})

	case *ForbiddenError:
		logger.Log.Warn().Err(err).Msg("Forbidden error")
		c.JSON(http.StatusForbidden, gin.H{"error": e})

	case *NotFoundError:
		logger.Log.Warn().Err(err).Msg("Not found error")
		c.JSON(http.StatusNotFound, gin.H{"error": e})

	case *ConflictError:
		logger.Log.Warn().Err(err).Msg("Conflict error")
		c.JSON(http.StatusConflict, gin.H{"error": e})

	case *PayloadTooLargeError:
		logger.Log.Warn().Err(err).Msg("Payload too large")
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": e})

	case *RateLimitError:
		logger.Log.Warn().Err(err).Msg("Rate limit exceeded")
		c.JSON(http.StatusTooManyRequests, gin.H{"error": e})

	default:
		logger.Log.Error().Err(err).Msg("Internal server error")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": &InternalError{Message: "An unexpected error occurred"},
		})
	}
}

func NewAuthError(message string) *AuthError {
	return &AuthError{
		Code:   "auth_error",
		Detail: message,
	}
}

func NewForbiddenError(message string) *ForbiddenError {
	return &ForbiddenError{
		Code:   "forbidden_error",
		Detail: message,
	}
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		Code:   "not_found_error",
		Detail: message,
	}
}

func NewConflictError(message string) *ConflictError {
	return &ConflictError{
		Code:   "conflict_error",
		Detail: message,
	}
}

func NewPayloadTooLargeError(message string) *PayloadTooLargeError {
	return &PayloadTooLargeError{
		Code:   "payload_too_large_error",
		Detail: message,
	}
}

func NewRateLimitError(message string) *RateLimitError {
	return &RateLimitError{
		Code:   "rate_limit_exceeded",
		Detail: message,
	}
}

func NewInternalError(message string) *InternalError {
	return &InternalError{
		Message: message,
	}
}

func NewValidationError(code, message string) *ValidationError {
	return &ValidationError{
		Code:   code,
		Detail: message,
	}
}
