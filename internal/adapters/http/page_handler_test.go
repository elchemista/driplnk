package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	handler "github.com/elchemista/driplnk/internal/adapters/http"
	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/mocks"
	"github.com/elchemista/driplnk/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestPageHandler_Dashboard(t *testing.T) {
	mockUsers := mocks.NewMockUserRepository()
	mockSessions := mocks.NewMockSessionManager()
	mockRepo := mocks.NewMockLinkRepository()
	mockAnalyticsRepo := mocks.NewMockAnalyticsRepository()
	mockMetadata := mocks.NewMockMetadataFetcher()

	linkService := service.NewLinkService(mockRepo, mockMetadata, nil)
	analyticsService := service.NewAnalyticsService(mockAnalyticsRepo)

	h := handler.NewPageHandler(mockUsers, mockSessions, linkService, analyticsService)

	t.Run("Success", func(t *testing.T) {
		user := &domain.User{ID: "user-1", Handle: "testuser", Theme: domain.Theme{}}
		mockUsers.AddUser(user)
		mockSessions.SetCurrentUser("user-1")

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		w := httptest.NewRecorder()

		h.Dashboard(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Dashboard")
	})

	t.Run("RedirectLogin", func(t *testing.T) {
		mockSessions.SetCurrentUser("") // Not logged in

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		w := httptest.NewRecorder()

		h.Dashboard(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "/login")
	})
}

func TestPageHandler_Profile(t *testing.T) {
	mockUsers := mocks.NewMockUserRepository()
	mockSessions := mocks.NewMockSessionManager()
	mockRepo := mocks.NewMockLinkRepository()
	mockAnalyticsRepo := mocks.NewMockAnalyticsRepository()
	mockMetadata := mocks.NewMockMetadataFetcher()

	linkService := service.NewLinkService(mockRepo, mockMetadata, nil)
	analyticsService := service.NewAnalyticsService(mockAnalyticsRepo)

	h := handler.NewPageHandler(mockUsers, mockSessions, linkService, analyticsService)

	t.Run("Success", func(t *testing.T) {
		user := &domain.User{ID: "user-1", Handle: "testuser", Theme: domain.Theme{}}
		mockUsers.AddUser(user)

		req := httptest.NewRequest(http.MethodGet, "/u/testuser", nil)
		req.SetPathValue("handle", "testuser")
		w := httptest.NewRecorder()

		h.Profile(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/unknown", nil)
		req.SetPathValue("handle", "unknown")
		w := httptest.NewRecorder()

		h.Profile(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
