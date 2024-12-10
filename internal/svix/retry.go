package svix

import (
	"errors"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/utils"
	svixapi "github.com/svix/svix-webhooks/go"
)

type retryConfig struct {
	attempts uint
	delay    time.Duration
}

var defaultRetryConfig = retryConfig{
	attempts: 5,
	delay:    time.Second,
}

// shouldRetry checks if a Svix error should be retried
func shouldRetry(err error) bool {
	var svixError *svixapi.Error
	if errors.As(err, &svixError) {
		shouldRetry := svixError.Status() == http.StatusTooManyRequests || // 429
			svixError.Status() >= http.StatusInternalServerError // All 5xx errors

		if shouldRetry {
			logger.Log.Debug().
				Int("status", svixError.Status()).
				Msg("Retriable Svix error encountered")
		}
		return shouldRetry
	}
	return false
}

// withRetry executes a Svix operation with retry logic
func withRetry(operation string, fn func() error) error {
	err := retry.Do(
		fn,
		retry.Attempts(defaultRetryConfig.attempts),
		retry.Delay(defaultRetryConfig.delay),
		retry.DelayType(retry.BackOffDelay),
		retry.RetryIf(shouldRetry),
		retry.OnRetry(func(n uint, err error) {
			logger.Log.Warn().
				Uint("attempt", n+1).
				Err(err).
				Str("operation", operation).
				Msg("Retrying Svix operation")
		}),
		// Add this option to preserve original error
		retry.LastErrorOnly(true),
	)

	// If it's already one of our custom error types, return it directly
	switch err.(type) {
	case *utils.ValidationError,
		*utils.AuthError,
		*utils.ForbiddenError,
		*utils.NotFoundError,
		*utils.ConflictError,
		*utils.PayloadTooLargeError,
		*utils.RateLimitError:
		return err
	}

	// Otherwise, check if it's a wrapped Svix error
	var retryErr *retry.Error
	if errors.As(err, &retryErr) {
		var svixErr *svixapi.Error
		if errors.As(retryErr.Unwrap(), &svixErr) {
			switch svixErr.Status() {
			case http.StatusBadRequest: // 400
				return utils.NewValidationError("invalid_request", svixErr.Error())
			case http.StatusUnauthorized: // 401
				return utils.NewAuthError(svixErr.Error())
			case http.StatusForbidden: // 403
				return utils.NewForbiddenError(svixErr.Error())
			case http.StatusNotFound: // 404
				return utils.NewNotFoundError(svixErr.Error())
			case http.StatusConflict: // 409
				return utils.NewConflictError("Event already processed (duplicate)")
			case http.StatusRequestEntityTooLarge: // 413
				return utils.NewPayloadTooLargeError(svixErr.Error())
			case http.StatusTooManyRequests: // 429
				return utils.NewRateLimitError(svixErr.Error())
			default:
				if svixErr.Status() >= 500 {
					return utils.NewInternalError("Svix service error")
				}
			}
		}
	}

	return err
}
