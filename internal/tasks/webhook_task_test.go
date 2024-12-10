package tasks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/markonick/gigs-challenge/internal/models"
)

type MockSvixClient struct {
	mock.Mock
}

func (m *MockSvixClient) SendMessage(ctx context.Context, appID string, event models.BaseEvent) error {
	args := m.Called(ctx, appID, event)
	return args.Error(0)
}

func (m *MockSvixClient) CreateApplication(ctx context.Context, name string) (string, error) {
	args := m.Called(ctx, name)
	return args.String(0), args.Error(1)
}

func (m *MockSvixClient) SetupApplicationEndpoints(ctx context.Context, appID string) error {
	args := m.Called(ctx, appID)
	return args.Error(0)
}

func TestWebhookTask_Execute(t *testing.T) {
	tests := []struct {
		name        string
		event       models.BaseEvent
		projectApps map[string]string
		setupMock   func(*MockSvixClient)
		wantErr     bool
	}{
		{
			name: "successful webhook delivery",
			event: models.BaseEvent{
				ID:      "event-123",
				Type:    "user.created",
				Project: "project-123",
			},
			projectApps: map[string]string{
				"project-123": "app-123",
			},
			setupMock: func(m *MockSvixClient) {
				m.On("SendMessage", mock.Anything, "app-123", mock.AnythingOfType("models.BaseEvent")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "missing project ID",
			event: models.BaseEvent{
				ID:   "event-123",
				Type: "user.created",
				// Project ID intentionally left empty
			},
			projectApps: map[string]string{},
			setupMock:   func(_ *MockSvixClient) {},
			wantErr:     true,
		},
		{
			name: "unknown project ID",
			event: models.BaseEvent{
				ID:      "event-123",
				Type:    "user.created",
				Project: "unknown-project",
			},
			projectApps: map[string]string{},
			setupMock:   func(_ *MockSvixClient) {},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockSvixClient)
			tt.setupMock(mockClient)

			task := NewWebhookTask(tt.event, mockClient, tt.projectApps)
			err := task.Execute(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockClient.AssertExpectations(t)
		})
	}
}
