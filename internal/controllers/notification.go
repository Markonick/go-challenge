package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		utils.RespondWithError(ctx, err)
		return
	}

	err = c.taskService.ProcessEvent(gigsEvent)
	if err != nil {
		utils.RespondWithError(ctx, err)
		return
	}

	ctx.Status(http.StatusAccepted)
}
