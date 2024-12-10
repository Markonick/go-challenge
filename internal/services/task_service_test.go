package services

import (
	"context"
	"testing"
	"time"

	"github.com/markonick/gigs-challenge/internal/models"
	"github.com/markonick/gigs-challenge/internal/utils"
	"github.com/markonick/gigs-challenge/internal/worker"
)

type MockTask struct {
	id    string
	event models.BaseEvent
	err   error
}

func (m *MockTask) Execute(_ context.Context) error {
	return m.err
}

func (m *MockTask) ID() string {
	return m.id
}

func TestTaskService_ProcessEvent(t *testing.T) {
	tests := []struct {
		name       string
		event      models.BaseEvent
		createTask func(models.BaseEvent) worker.Task
		wantErr    bool
	}{
		{
			name: "successfully submits task",
			event: models.BaseEvent{
				ID:   "evt_123",
				Data: map[string]any{"id": "123"},
			},
			createTask: func(event models.BaseEvent) worker.Task {
				return &MockTask{id: event.ID, event: event}
			},
			wantErr: false,
		},
		{
			name: "handles task creation",
			event: models.BaseEvent{
				ID:      "evt_123",
				Type:    "test.event",
				Project: "test",
				Data:    map[string]any{"id": "123"},
			},
			createTask: func(event models.BaseEvent) worker.Task {
				return &MockTask{
					id: event.ID,
					err: &utils.ValidationError{
						Code:   "validation_error",
						Detail: "bad request",
					},
				}
			},
			wantErr: false, // Task submission should succeed even if task will fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service with small worker pool
			service := NewTaskService(2, tt.createTask)

			// Submit task
			err := service.ProcessEvent(tt.event)

			// Verify submission result
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessEvent() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Give workers time to process
			time.Sleep(100 * time.Millisecond)
		})
	}
}
