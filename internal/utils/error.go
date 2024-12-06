package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/markonick/gigs-challenge/internal/logger"
)

func HandleError(c *gin.Context, statusCode int, err error, message string) {
	logger.Log.Error().Err(err).Msg(message)
	c.JSON(statusCode, gin.H{"error": message})
}
