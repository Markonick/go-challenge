package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/markonick/gigs-challenge/internal/models"
	"github.com/markonick/gigs-challenge/internal/utils"
)

func ParsePubSubMessage(c *gin.Context) (models.BaseEvent, error) {
	var gigsEvent models.BaseEvent
	if err := c.ShouldBindJSON(&gigsEvent); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Return the first validation error
			for _, err := range validationErrors {
				return models.BaseEvent{}, &utils.ValidationError{
					Code:   err.Field(),
					Detail: utils.GetValidationMessage(err.Tag()),
				}
			}
		}
		// If it's not a validation error but JSON binding still failed
		// (e.g., malformed JSON), return a generic validation error
		return models.BaseEvent{}, &utils.ValidationError{
			Code:   "body",
			Detail: "Invalid JSON format in request body",
		}
	}

	return gigsEvent, nil
}
