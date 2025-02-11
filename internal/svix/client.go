package svix

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/models"
	"github.com/markonick/gigs-challenge/internal/utils"
	svixapi "github.com/svix/svix-webhooks/go"
)

type Client interface {
	CreateApplication(ctx context.Context, name string) (string, error)
	SetupApplicationEndpoints(ctx context.Context, appID string) error
	SendMessage(ctx context.Context, appID string, event models.BaseEvent) error
}

type clientImpl struct {
	svix *svixapi.Svix
}

func NewClient(svixToken string) Client {
	// Let Svix infer the server URL from the token
	return &clientImpl{
		svix: svixapi.New(svixToken, nil),
	}
}

func (c *clientImpl) CreateApplication(ctx context.Context, name string) (string, error) {
	// First check if application already exists
	apps, err := c.svix.Application.List(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list applications: %w", err)
	}

	// Check for existing app with same name
	for _, app := range apps.Data {
		if app.Name == name {
			logger.Log.Info().
				Str("app_id", app.Id).
				Str("name", name).
				Msg("Found existing Svix application")
			return app.Id, nil
		}
	}

	// If not found, create new application
	var appID string
	err = withRetry("create_application", func() error {
		rateLimit := int32(1)
		app, err := c.svix.Application.Create(ctx, &svixapi.ApplicationIn{
			Name:      name,
			RateLimit: *svixapi.NullableInt32(&rateLimit)})
		if err != nil {
			return err
		}
		appID = app.Id
		return nil
	})
	return appID, err
}

func (c *clientImpl) SetupApplicationEndpoints(ctx context.Context, appID string) error {
	// First create all event types
	for _, eventType := range models.GetCommonEventTypes() {
		eventTypeStr := string(eventType)

		err := withRetry("create_event_type", func() error {
			eventTypeIn := &svixapi.EventTypeIn{
				Name:        eventTypeStr,
				Description: fmt.Sprintf("Event type for %s", eventTypeStr),
			}

			_, err := c.svix.EventType.Create(ctx, eventTypeIn)
			if err != nil {
				if apiErr, ok := err.(*svixapi.Error); ok && apiErr.Status() == 409 {
					logger.Log.Debug().
						Str("event_type", eventTypeStr).
						Msg("Event type exists")
					return nil
				}

				logger.Log.Error().
					Str("event_type", eventTypeStr).
					Err(err).
					Msg("Failed to create event type")
				return fmt.Errorf("failed to create event type: %w", err)
			}

			logger.Log.Info().
				Str("event_type", eventTypeStr).
				Msg("Created new event type")
			return nil
		})

		if err != nil {
			return err
		}
	}

	// Then list all existing endpoints
	endpoints, err := c.svix.Endpoint.List(ctx, appID, nil)
	if err != nil {
		return fmt.Errorf("failed to list endpoints: %w", err)
	}

	for _, eventType := range models.GetCommonEventTypes() {
		endpointURL := fmt.Sprintf("https://your-api-domain.com/webhooks/%s", eventType)

		// Check if endpoint already exists
		endpointExists := false
		for _, ep := range endpoints.Data {
			if ep.Url == endpointURL {
				logger.Log.Info().
					Str("event_type", string(eventType)).
					Str("app_id", appID).
					Str("endpoint", endpointURL).
					Msg("Endpoint already exists")
				endpointExists = true
				break
			}
		}

		if !endpointExists {
			err := withRetry("create_endpoint", func() error {
				version := int32(1)
				endpointIn := &svixapi.EndpointIn{
					Url:         endpointURL,
					Description: svixapi.String(fmt.Sprintf("Endpoint for %s events", eventType)),
					FilterTypes: []string{string(eventType)},
					Version:     *svixapi.NullableInt32(&version), // Add dereferencing operator *
				}

				_, err := c.svix.Endpoint.Create(ctx, appID, endpointIn)
				if err != nil {
					if apiErr, ok := err.(*svixapi.Error); ok {
						logger.Log.Error().
							Str("event_type", string(eventType)).
							Int("status", apiErr.Status()).
							Str("message", apiErr.Error()).
							Msg("Failed to create endpoint")
					}
					return err
				}
				logger.Log.Info().
					Str("event_type", string(eventType)).
					Str("app_id", appID).
					Str("endpoint", endpointURL).
					Msg("Created new endpoint")
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to create endpoint for %s: %w", eventType, err)
			}
		}
	}
	return nil
}

func (c *clientImpl) SendMessage(ctx context.Context, appID string, event models.BaseEvent) error {
	message := &svixapi.MessageIn{
		EventId:   *svixapi.NullableString(&event.ID),
		EventType: event.Type,
		Payload:   event.Data,
	}

	err := withRetry("send_message", func() error {
		_, err := c.svix.Message.Create(ctx, appID, message)
		if err != nil {
			logger.Log.Debug().
				Str("error_type", fmt.Sprintf("%T", err)).
				Msg("Error from Svix API")

			var svixError *svixapi.Error
			if errors.As(err, &svixError) {
				switch svixError.Status() {
				case http.StatusConflict: // 409
					return utils.NewConflictError(svixError.Error())
				case http.StatusUnauthorized: // 401
					return utils.NewAuthError(svixError.Error())
				case http.StatusForbidden: // 403
					return utils.NewForbiddenError(svixError.Error())
				case http.StatusNotFound: // 404
					return utils.NewNotFoundError(svixError.Error())
				case http.StatusRequestEntityTooLarge: // 413
					return utils.NewPayloadTooLargeError(svixError.Error())
				case http.StatusTooManyRequests: // 429
					return utils.NewRateLimitError(svixError.Error())
				case http.StatusUnprocessableEntity: // 422
					return utils.NewValidationError("validation_failed", svixError.Error())
				default:
					if svixError.Status() >= 500 {
						return utils.NewInternalError(svixError.Error())
					}
					return err
				}
			}
			return err
		}
		return nil
	})

	if err != nil {
		logger.Log.Debug().
			Str("error_type", fmt.Sprintf("%T", err)).
			Msg("Error after retry")
	}
	return err
}
