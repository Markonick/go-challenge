package models

import (
	"github.com/go-playground/validator/v10"
)

// BaseEvent reprsents the common fields for Gigs events
// The struct tags are used by:
// encoding/json package uses json:"id" for marshaling/unmarshaling
// Gin's validator uses binding:"required" for validation
type BaseEvent struct {
	ID      string                 `json:"id" binding:"required"`
	Type    string                 `json:"type" binding:"required"`
	Project string                 `json:"project" binding:"required"`
	Data    map[string]interface{} `json:"data" binding:"required"`
}

// RegisterValidators registers custom validators for BaseEvent
func RegisterValidators(v *validator.Validate) error {
	if err := v.RegisterValidation("eventIDFormat", validateEventID); err != nil {
		return err
	}
	if err := v.RegisterValidation("validEventType", validateEventType); err != nil {
		return err
	}
	if err := v.RegisterValidation("validProject", validateProject); err != nil {
		return err
	}
	return nil
}

func validateEventID(fl validator.FieldLevel) bool {
	id := fl.Field().String()
	return len(id) > 4 && id[:4] == "evt_"
}

func validateEventType(fl validator.FieldLevel) bool {
	eventType := EventType(fl.Field().String())
	validTypes := GetCommonEventTypes()
	for _, t := range validTypes {
		if t == eventType {
			return true
		}
	}
	return false
}

func validateProject(fl validator.FieldLevel) bool {
	project := fl.Field().String()
	validProjects := []string{"dev", "staging", "prod"}
	for _, p := range validProjects {
		if p == project {
			return true
		}
	}
	return false
}
