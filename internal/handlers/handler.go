package handlers

import (
	"context"

	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/models"
	"github.com/markonick/gigs-challenge/internal/svix"
)

type Handler struct {
	svixClient svix.Client
}

func NewHandler(svixClient svix.Client) *Handler {
	return &Handler{
		svixClient: svixClient,
	}
}

func (h *Handler) Handle(ctx context.Context, event models.BaseEvent) error {
	logger.Log.Info().
		Str("type", event.Type).
		Str("eventID", event.ID).
		Interface("data", event.Data).
		Msg("Received and handling event")

	appID, err := h.svixClient.CreateApplication(ctx, event.Type)
	if err != nil {
		return err
	}

	return h.svixClient.SendMessage(ctx, appID, event)
}
