package http_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	adapters_http "github.com/elchemista/driplnk/internal/adapters/http"
	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/mocks"
	"github.com/elchemista/driplnk/internal/service"
)

func TestLinkHandler_CreateLink(t *testing.T) {
	userRepo := mocks.NewMockUserRepository()
	linkRepo := mocks.NewMockLinkRepository()
	analyticsRepo := mocks.NewMockAnalyticsRepository()
	sessionMgr := mocks.NewMockSessionManager()

	linkSvc := service.NewLinkService(linkRepo)
	analyticsSvc := service.NewAnalyticsService(analyticsRepo)
	handler := adapters_http.NewLinkHandler(linkSvc, analyticsSvc, sessionMgr, userRepo)

	// Setup user
	user := &domain.User{
		ID:     "user-123",
		Email:  "test@example.com",
		Handle: "testuser",
	}
	userRepo.AddUser(user)
	sessionMgr.SetCurrentUser("user-123")

	t.Run("creates link with valid data", func(t *testing.T) {
		form := url.Values{}
		form.Set("title", "My Website")
		form.Set("url", "https://example.com")
		form.Set("type", "standard")

		req := httptest.NewRequest(http.MethodPost, "/dashboard/links", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.CreateLink(rec, req)

		if rec.Code != http.StatusSeeOther && rec.Code != http.StatusOK {
			t.Errorf("expected redirect or OK, got %d", rec.Code)
		}
	})

	t.Run("returns error for missing title", func(t *testing.T) {
		form := url.Values{}
		form.Set("url", "https://example.com")

		req := httptest.NewRequest(http.MethodPost, "/dashboard/links", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.CreateLink(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("returns 401 for unauthenticated user", func(t *testing.T) {
		sessionMgr.SetCurrentUser("") // Clear session

		form := url.Values{}
		form.Set("title", "My Website")
		form.Set("url", "https://example.com")

		req := httptest.NewRequest(http.MethodPost, "/dashboard/links", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.CreateLink(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}

		sessionMgr.SetCurrentUser("user-123") // Restore for other tests
	})

	t.Run("responds with Turbo Stream for Turbo request", func(t *testing.T) {
		form := url.Values{}
		form.Set("title", "Turbo Link")
		form.Set("url", "https://turbo.example.com")
		form.Set("type", "standard")

		req := httptest.NewRequest(http.MethodPost, "/dashboard/links", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "text/vnd.turbo-stream.html")
		rec := httptest.NewRecorder()

		handler.CreateLink(rec, req)

		contentType := rec.Header().Get("Content-Type")
		if !strings.Contains(contentType, "turbo-stream") {
			t.Errorf("expected turbo-stream content type, got %s", contentType)
		}
		if !strings.Contains(rec.Body.String(), "<turbo-stream") {
			t.Error("expected turbo-stream in response body")
		}
	})
}

func TestLinkHandler_DeleteLink(t *testing.T) {
	userRepo := mocks.NewMockUserRepository()
	linkRepo := mocks.NewMockLinkRepository()
	analyticsRepo := mocks.NewMockAnalyticsRepository()
	sessionMgr := mocks.NewMockSessionManager()

	linkSvc := service.NewLinkService(linkRepo)
	analyticsSvc := service.NewAnalyticsService(analyticsRepo)
	handler := adapters_http.NewLinkHandler(linkSvc, analyticsSvc, sessionMgr, userRepo)

	user := &domain.User{
		ID:     "user-123",
		Email:  "test@example.com",
		Handle: "testuser",
	}
	userRepo.AddUser(user)
	sessionMgr.SetCurrentUser("user-123")

	linkRepo.AddLink(&domain.Link{
		ID:        "link-to-delete",
		UserID:    "user-123",
		Title:     "Test Link",
		URL:       "https://example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	t.Run("deletes link successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/dashboard/links/link-to-delete/delete", nil)
		req.SetPathValue("id", "link-to-delete")
		rec := httptest.NewRecorder()

		handler.DeleteLink(rec, req)

		if rec.Code != http.StatusSeeOther && rec.Code != http.StatusOK {
			t.Errorf("expected redirect or OK, got %d", rec.Code)
		}
	})

	t.Run("returns 401 for unauthenticated", func(t *testing.T) {
		sessionMgr.SetCurrentUser("")
		req := httptest.NewRequest(http.MethodPost, "/dashboard/links/some-link/delete", nil)
		req.SetPathValue("id", "some-link")
		rec := httptest.NewRecorder()

		handler.DeleteLink(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestLinkHandler_HandleRedirect(t *testing.T) {
	userRepo := mocks.NewMockUserRepository()
	linkRepo := mocks.NewMockLinkRepository()
	analyticsRepo := mocks.NewMockAnalyticsRepository()
	sessionMgr := mocks.NewMockSessionManager()

	linkSvc := service.NewLinkService(linkRepo)
	analyticsSvc := service.NewAnalyticsService(analyticsRepo)
	handler := adapters_http.NewLinkHandler(linkSvc, analyticsSvc, sessionMgr, userRepo)

	linkRepo.AddLink(&domain.Link{
		ID:       "active-link",
		UserID:   "user-123",
		URL:      "https://target.example.com",
		IsActive: true,
	})

	linkRepo.AddLink(&domain.Link{
		ID:       "inactive-link",
		UserID:   "user-123",
		URL:      "https://inactive.example.com",
		IsActive: false,
	})

	t.Run("redirects to target URL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/go/active-link", nil)
		req.SetPathValue("id", "active-link")
		rec := httptest.NewRecorder()

		handler.HandleRedirect(rec, req)

		if rec.Code != http.StatusTemporaryRedirect {
			t.Errorf("expected 307, got %d", rec.Code)
		}
		location := rec.Header().Get("Location")
		if location != "https://target.example.com" {
			t.Errorf("expected redirect to target URL, got %s", location)
		}
	})

	t.Run("returns 404 for inactive link", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/go/inactive-link", nil)
		req.SetPathValue("id", "inactive-link")
		rec := httptest.NewRecorder()

		handler.HandleRedirect(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("returns 404 for non-existent link", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/go/unknown", nil)
		req.SetPathValue("id", "unknown")
		rec := httptest.NewRecorder()

		handler.HandleRedirect(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}
