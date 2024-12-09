package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/markonick/gigs-challenge/internal/models"
)

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) ProcessEvent(event models.BaseEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

var tests = []struct {
	name           string
	requestBody    string
	setupMock      func(*MockTaskService)
	expectedStatus int
}{
	{
		name: "successful event processing",
		requestBody: `{
			"id": "evt_123",
			"type": "test.event",
			"project": "test",
			"created_at": "2024-01-01T00:00:00Z",
			"data": {"id": "123", "name": "Test Project"}
		}`,
		setupMock: func(m *MockTaskService) {
			m.On("ProcessEvent", mock.MatchedBy(func(event models.BaseEvent) bool {
				return event.ID == "evt_123" && event.Type == "test.event" && event.Project == "test"
			})).Return(nil)
		},
		expectedStatus: http.StatusAccepted,
	},
	{
		name:        "invalid JSON payload",
		requestBody: `{"invalid": json}`,
		setupMock: func(_ *MockTaskService) {
			// No mock expectations - should fail before service call
		},
		expectedStatus: http.StatusBadRequest,
	},
	{
		name: "missing required fields",
		requestBody: `{
			"id": "evt_123"
		}`,
		setupMock: func(_ *MockTaskService) {
			// No mock expectations - should fail validation
		},
		expectedStatus: http.StatusBadRequest,
	},
	{
		name: "service error",
		requestBody: `{
			"id": "evt_123",
			"type": "test.event",
			"project": "test",
			"created_at": "2024-01-01T00:00:00Z",
			"data": {"id": "123"}
		}`,
		setupMock: func(m *MockTaskService) {
			m.On("ProcessEvent", mock.Anything).Return(assert.AnError)
		},
		expectedStatus: http.StatusInternalServerError,
	},
}

func TestNotificationController_Create(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//Setup
			gin.SetMode(gin.TestMode)
			mockTaskService := new(MockTaskService)
			test.setupMock(mockTaskService)

			controller := NewNotificationController(mockTaskService)

			// Record the response
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			ctx.Request = httptest.NewRequest(http.MethodPost, "/notifications", strings.NewReader(test.requestBody))
			ctx.Request.Header.Set("Content-Type", "application/json")

			// Execute
			controller.Create(ctx)

			// Wait for response to be written
			ctx.Writer.Flush()

			// Debug output
			t.Logf("Response Status: %d", w.Code)
			t.Logf("Response Body: %s", w.Body.String())
			t.Logf("Response Headers: %v", w.Header())

			// Assert
			assert.Equal(t, test.expectedStatus, w.Code)
			// mockTaskService.AssertExpectations(t)
		})
	}
}
