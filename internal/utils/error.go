package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/markonick/gigs-challenge/internal/logger"
)

// ValidationError represents a structured validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type rateLimitError struct{}

func (r *rateLimitError) Error() string {
	return "Rate limit exceeded. Please try again later."
}

// Error implements the error interface
func (v *ValidationError) Error() string {
	return v.Message
}

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

func HandleError(c *gin.Context, statusCode int, err error, message string) {
	logger.Log.Error().Err(err).Msg(message)

	// Check for rate limit error first
	if _, ok := err.(*rateLimitError); ok {
		c.JSON(statusCode, gin.H{
			"error": "Rate limit exceeded. Please try again later.",
		})
		return
	}
	if validationErr, ok := err.(*ValidationError); ok {
		c.JSON(statusCode, gin.H{
			"errors": []ValidationError{*validationErr},
		})
		return
	}

	c.JSON(statusCode, gin.H{"error": message})
}

type SvixErrorResponse struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}
