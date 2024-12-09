package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/services"
	"github.com/markonick/gigs-challenge/internal/utils"
)

type NotificationController struct {
	taskService services.TaskService
}

func NewNotificationController(taskService services.TaskService) *NotificationController {
	return &NotificationController{
		taskService: taskService,
	}
}

func (c *NotificationController) Create(ctx *gin.Context) {
	gigsEvent, err := ParsePubSubMessage(ctx)
	if err != nil {
		if validationErr, ok := err.(*utils.ValidationError); ok {
			utils.HandleError(ctx, http.StatusBadRequest, validationErr, validationErr.Message)
		} else {
			utils.HandleError(ctx, http.StatusBadRequest, err, "Failed to parse Pub/Sub message")
		}
		return
	}

	logger.Log.Info().
		Str("event_type", string(gigsEvent.Type)).
		Str("event_id", gigsEvent.ID).
		Msg("Received event")

	// ProcessEvent is asynchronous
	c.taskService.ProcessEvent(gigsEvent)

	logger.Log.Info().
		Str("event_id", gigsEvent.ID).
		Msg("Task submitted to worker pool")

	// Immediately respond with 202 Accepted
	ctx.Status(http.StatusAccepted)
}
