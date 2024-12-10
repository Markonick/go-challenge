package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"github.com/markonick/gigs-challenge/internal/models"
	"github.com/markonick/gigs-challenge/internal/utils"
)

const baseRequestBody = `{
    "id": "evt_123",
    "type": "test.event",
    "project": "test",
    "data": {"id": "123"}
}`

type MockTaskService struct {
	mock.Mock
}

// Simplified interface - no more channels
func (m *MockTaskService) ProcessEvent(event models.BaseEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

var tests = []struct {
	name        string
	requestBody string
	setupMock   func(*MockTaskService)
}{
	{
		name:        "successful event processing",
		requestBody: baseRequestBody,
		setupMock: func(m *MockTaskService) {
			m.On("ProcessEvent", mock.MatchedBy(func(event models.BaseEvent) bool {
				return event.Type == "test.event" && event.Project == "test"
			})).Return(nil)
		},
	},
	{
		name:        "rate limit exceeded",
		requestBody: baseRequestBody,
		setupMock: func(m *MockTaskService) {
			m.On("ProcessEvent", mock.Anything).Return(
				utils.NewRateLimitError("Rate limit exceeded"),
			)
		},
	},
	// ... other error cases remain similar, just remove channel handling
	{
		name:        "invalid JSON payload",
		requestBody: `{"invalid": json}`,
		setupMock: func(_ *MockTaskService) {
			// No mock expectations - should fail before service call
		},
	},
}

func TestNotificationController_Create(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			mockTaskService := new(MockTaskService)
			test.setupMock(mockTaskService)

			controller := NewNotificationController(mockTaskService)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			ctx.Request = httptest.NewRequest(
				http.MethodPost,
				"/notifications",
				strings.NewReader(test.requestBody),
			)
			ctx.Request.Header.Set("Content-Type", "application/json")

			controller.Create(ctx)

			mockTaskService.AssertExpectations(t)
		})
	}
}
