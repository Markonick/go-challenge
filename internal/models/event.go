package models

import "time"

// BaseEvent reprsents the common fields for Gigs events
type BaseEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	CreatedAt time.Time              `json:"created_at"`
	Data      map[string]interface{} `json:"data"`
}
