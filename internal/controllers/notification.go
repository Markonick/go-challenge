package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markonick/gigs-challenge/internal/services"
	"github.com/markonick/gigs-challenge/internal/utils"
)

type NotificationResponse struct {
	TaskID  string      `json:"task_id,omitempty"`
	EventID string      `json:"event_id,omitempty"`
	Status  string      `json:"status"`
	Error   string      `json:"error,omitempty"`
	Details interface{} `json:"details,omitempty"`
}
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
		utils.RespondWithError(ctx, err)
		return
	}

	err = c.taskService.ProcessEvent(gigsEvent)
	if err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	// Success response
	ctx.JSON(http.StatusAccepted, NotificationResponse{
		TaskID:  gigsEvent.ID,
		EventID: gigsEvent.ID,
		Status:  "accepted",
	})
}
