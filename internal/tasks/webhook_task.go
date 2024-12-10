package tasks

import (
	"context"
	"fmt"

	svixapi "github.com/svix/svix-webhooks/go"

	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/models"
	"github.com/markonick/gigs-challenge/internal/svix"
	"github.com/markonick/gigs-challenge/internal/utils"
)

// WebhookTask implements worker.Task interface
type WebhookTask struct {
	event         models.BaseEvent
	svixClient    svix.Client
	projectAppIDs map[string]string
}

// NewWebhookTask creates a new webhook task that implements worker.Task
func NewWebhookTask(event models.BaseEvent, svixClient svix.Client, projectAppIDs map[string]string) *WebhookTask {
	return &WebhookTask{
		event:         event,
		svixClient:    svixClient,
		projectAppIDs: projectAppIDs,
	}
}

// Process implements worker.Task interface
func (t *WebhookTask) Execute(ctx context.Context) error {
	projectID := t.event.Project
	if projectID == "" {
		return fmt.Errorf("project not found in event data")
	}

	appID, ok := t.projectAppIDs[projectID]
	if !ok {
		return fmt.Errorf("no app ID found for project: %s", projectID)
	}

	logger.Log.Info().
		Str("type", t.event.Type).
		Str("eventID", t.event.ID).
		Msg("Processing webhook event")

	err := t.svixClient.SendMessage(ctx, appID, t.event)
	if err != nil {
		// Check if it's a Svix API error
		if apiErr, ok := err.(svixapi.Error); ok {
			if apiErr.Status() == 409 {
				logger.Log.Info().
					Str("event_id", t.event.ID).
					Str("type", string(t.event.Type)).
					Msg("Event already processed (duplicate)")
				return nil // Don't treat duplicates as errors
			}
			// Other Svix API errors
			return &utils.ConflictError{
				Code:   apiErr.Error(),
				Detail: string(apiErr.Body()),
			}
		}
		// Unexpected errors
		return fmt.Errorf("webhook delivery failed: %w", err)
	}

	return nil
}

// ID implements worker.Task interface
func (t *WebhookTask) ID() string {
	return t.event.ID
}
