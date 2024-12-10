package svix

import (
	"errors"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/markonick/gigs-challenge/internal/logger"
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
	return err
}
