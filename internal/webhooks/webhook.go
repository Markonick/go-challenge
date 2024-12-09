package webhook

import (
	"context"
	"fmt"

	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/models"
	"github.com/markonick/gigs-challenge/internal/svix"
)

type WebhookTask struct {
	event         models.BaseEvent
	svixClient    svix.Client
	projectAppIDs map[string]string
}

func NewWebhookTask(event models.BaseEvent, svixClient svix.Client, projectAppIDs map[string]string) *WebhookTask {
	return &WebhookTask{
		event:         event,
		svixClient:    svixClient,
		projectAppIDs: projectAppIDs,
	}
}

func (t *WebhookTask) Process(ctx context.Context) error {
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

	return t.svixClient.SendMessage(ctx, appID, t.event)
}

func (t *WebhookTask) ID() string {
	return t.event.ID
}
