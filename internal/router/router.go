package router

import (
	"github.com/gin-gonic/gin"
	controller "github.com/markonick/gigs-challenge/internal/controllers"
)

func Setup(notificationCtrl *controller.NotificationController) *gin.Engine {
	r := gin.Default()
	r.POST("/notifications", notificationCtrl.Create)
	return r
}
