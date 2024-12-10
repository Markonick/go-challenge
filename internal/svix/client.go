package svix

import (
	"context"
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
		rateLimit := int32(1000)
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
					logger.Log.Debug(). // Reduced to DEBUG level
								Str("event_type", eventTypeStr).
								Msg("Event type exists")
					return nil
				}
				// Only log real errors as ERROR
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

			if apiErr, ok := err.(*svixapi.Error); ok && apiErr.Status() == 409 {
				logger.Log.Debug().
					Str("event_id", event.ID).
					Str("event_type", string(event.Type)).
					Str("app_id", appID).
					Msg("Skipping duplicate event")
				switch apiErr.Status() {
				case http.StatusBadRequest: // 400
					return &utils.ValidationError{
						Code:   "invalid_request",
						Detail: apiErr.Error(),
					}
				case http.StatusUnauthorized: // 401
					return &utils.AuthError{
						Code:   "unauthorized",
						Detail: apiErr.Error(),
					}
				case http.StatusForbidden: // 403
					return &utils.ForbiddenError{
						Code:   "forbidden",
						Detail: apiErr.Error(),
					}
				case http.StatusNotFound: // 404
					return &utils.NotFoundError{
						Code:   "not_found",
						Detail: apiErr.Error(),
					}
				case http.StatusConflict: // 409
					logger.Log.Info().
						Str("event_id", event.ID).
						Msg("Event already processed (duplicate)")
					return nil // Don't treat 409 as error
				case http.StatusRequestEntityTooLarge: // 413
					return &utils.PayloadTooLargeError{
						Code:   "payload_too_large",
						Detail: apiErr.Error(),
					}
				case http.StatusTooManyRequests: // 429
					return &utils.RateLimitError{
						Code:   "rate_limit_exceeded",
						Detail: apiErr.Error(),
					}
				default:
					if apiErr.Status() >= 500 {
						return &utils.InternalError{
							Message: "Svix service error",
						}
					}
					return &utils.InternalError{
						Message: apiErr.Error(),
					}
				}
			}
			return err
		}
		return nil
	})

	if err != nil {
		logger.Log.Error().
			Str("app_id", appID).
			Err(err).
			Msg("Failed to send message to Svix after retries")
		return err
	}

	logger.Log.Info().
		Str("app_id", appID).
		Str("event_id", event.ID).
		Str("event_type", string(event.Type)).
		Msg("Message sent to Svix successfully")

	return nil
}
