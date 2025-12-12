package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	handler "github.com/elchemista/driplnk/internal/adapters/http"
	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/mocks"
	"github.com/elchemista/driplnk/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestLinkHandler_CreateLink(t *testing.T) {
	mockRepo := mocks.NewMockLinkRepository()
	mockAnalyticsRepo := mocks.NewMockAnalyticsRepository()
	mockSessionManager := mocks.NewMockSessionManager()
	mockUserRepo := mocks.NewMockUserRepository()
	mockMetadata := mocks.NewMockMetadataFetcher()

	linkService := service.NewLinkService(mockRepo, mockMetadata, nil)
	analyticsService := service.NewAnalyticsService(mockAnalyticsRepo)
	h := handler.NewLinkHandler(linkService, analyticsService, mockSessionManager, mockUserRepo)

	t.Run("Success", func(t *testing.T) {
		// Seed User and Session
		user := &domain.User{ID: "user-123"}
		mockUserRepo.AddUser(user)
		mockSessionManager.SetCurrentUser("user-123")

		// Request
		form := url.Values{}
		form.Add("title", "My Link")
		form.Add("url", "https://example.com")
		form.Add("type", "standard")
		req := httptest.NewRequest(http.MethodPost, "/dashboard/links", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.CreateLink(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)

		// Verify creation
		links, _ := mockRepo.ListByUser(context.Background(), "user-123")
		assert.Len(t, links, 1)
		assert.Equal(t, "My Link", links[0].Title)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		mockSessionManager.SetCurrentUser("") // Logged out

		req := httptest.NewRequest(http.MethodPost, "/dashboard/links", nil)
		w := httptest.NewRecorder()

		h.CreateLink(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestLinkHandler_HandleRedirect(t *testing.T) {
	mockRepo := mocks.NewMockLinkRepository()
	mockAnalyticsRepo := mocks.NewMockAnalyticsRepository()
	mockSessionManager := mocks.NewMockSessionManager()
	mockUserRepo := mocks.NewMockUserRepository()
	mockMetadata := mocks.NewMockMetadataFetcher()

	linkService := service.NewLinkService(mockRepo, mockMetadata, nil)
	analyticsService := service.NewAnalyticsService(mockAnalyticsRepo)
	h := handler.NewLinkHandler(linkService, analyticsService, mockSessionManager, mockUserRepo)

	t.Run("Success", func(t *testing.T) {
		linkID := domain.LinkID("link-123")
		link := &domain.Link{
			ID:       linkID,
			UserID:   "user-1",
			URL:      "https://destination.com",
			IsActive: true,
		}
		mockRepo.AddLink(link)

		req := httptest.NewRequest(http.MethodGet, "/go/link-123", nil)
		req.SetPathValue("id", "link-123")
		w := httptest.NewRecorder()

		h.HandleRedirect(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, "https://destination.com", w.Header().Get("Location"))

		// Note: Async tracking is hard to test here without a sync mechanism or wait,
		// but we trust the service call happens.
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/go/invalid", nil)
		req.SetPathValue("id", "invalid")
		w := httptest.NewRecorder()

		h.HandleRedirect(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestLinkHandler_DeleteLink(t *testing.T) {
	mockRepo := mocks.NewMockLinkRepository()
	mockSessionManager := mocks.NewMockSessionManager()
	mockUserRepo := mocks.NewMockUserRepository()
	mockMetadata := mocks.NewMockMetadataFetcher()

	linkService := service.NewLinkService(mockRepo, mockMetadata, nil)
	h := handler.NewLinkHandler(linkService, nil, mockSessionManager, mockUserRepo)

	t.Run("Success", func(t *testing.T) {
		user := &domain.User{ID: "user-1"}
		mockUserRepo.AddUser(user)
		mockSessionManager.SetCurrentUser("user-1")

		link := &domain.Link{ID: "link-1", UserID: "user-1"}
		mockRepo.AddLink(link)

		req := httptest.NewRequest(http.MethodPost, "/dashboard/links/link-1/delete", nil)
		req.SetPathValue("id", "link-1")
		w := httptest.NewRecorder()

		h.DeleteLink(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)

		// Verify
		_, err := mockRepo.GetByID(context.Background(), "link-1")
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}
