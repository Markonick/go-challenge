package svix

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/models"
	svixapi "github.com/svix/svix-webhooks/go"
)

type Client interface {
	CreateApplication(ctx context.Context, name string) (string, error)
	SendMessage(ctx context.Context, appID string, event models.BaseEvent) error
}

type clientImpl struct {
	svix *svixapi.Svix
}

func NewClient(apiKey string) Client {
	return &clientImpl{
		svix: svixapi.New(apiKey, nil),
	}
}

func shouldRetry(err error) bool {
	var svixError *svixapi.Error
	if errors.As(err, &svixError) {
		return svixError.Status() == http.StatusTooManyRequests || // 429
			svixError.Status() >= http.StatusInternalServerError // All 5xx errors
	}
	return false
}

func (c *clientImpl) withRetry(operation string, fn func() error) error {
	return retry.Do(
		fn,
		retry.Attempts(5),
		retry.Delay(time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.RetryIf(shouldRetry),
		retry.OnRetry(func(n uint, err error) {
			logger.Log.Warn().
				Uint("attempt", n+1).
				Err(err).
				Str("operation", operation).
				Msg("Retrying Svix operation")
		}),
	)
}

func (c *clientImpl) CreateApplication(ctx context.Context, name string) (string, error) {
	var appID string
	err := c.withRetry("create_application", func() error {
		app, err := c.svix.Application.Create(ctx, &svixapi.ApplicationIn{Name: name})
		if err != nil {
			return err
		}
		appID = app.Id
		return nil
	})
	return appID, err
}

func (c *clientImpl) SendMessage(ctx context.Context, appID string, event models.BaseEvent) error {
	message := &svixapi.MessageIn{
		EventId:   *svixapi.NullableString(&event.ID),
		EventType: event.Type,
		Payload:   event.Data,
	}

	return c.withRetry("send_message", func() error {
		_, err := c.svix.Message.Create(ctx, appID, message)
		return err
	})
}
