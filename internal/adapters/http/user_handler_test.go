package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	adapters_http "github.com/elchemista/driplnk/internal/adapters/http"
	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/mocks"
)

func TestUserHandler_UpdateProfile(t *testing.T) {
	userRepo := mocks.NewMockUserRepository()
	sessionMgr := mocks.NewMockSessionManager()
	handler := adapters_http.NewUserHandler(userRepo, sessionMgr)

	user := &domain.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Handle:    "testuser",
		Title:     "Original Title",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	userRepo.AddUser(user)
	sessionMgr.SetCurrentUser("user-123")

	t.Run("updates profile successfully", func(t *testing.T) {
		form := url.Values{}
		form.Set("title", "New Title")
		form.Set("handle", "testuser")
		form.Set("description", "New description")

		req := httptest.NewRequest(http.MethodPost, "/dashboard/profile", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.UpdateProfile(rec, req)

		if rec.Code != http.StatusSeeOther && rec.Code != http.StatusOK {
			t.Errorf("expected redirect or OK, got %d", rec.Code)
		}

		// Verify update
		updated, _ := userRepo.GetByID(context.Background(), "user-123")
		if updated.Title != "New Title" {
			t.Errorf("expected title 'New Title', got '%s'", updated.Title)
		}
	})

	t.Run("returns 401 for unauthenticated", func(t *testing.T) {
		sessionMgr.SetCurrentUser("")
		form := url.Values{}
		form.Set("title", "Hacked")

		req := httptest.NewRequest(http.MethodPost, "/dashboard/profile", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.UpdateProfile(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}

		sessionMgr.SetCurrentUser("user-123")
	})

	t.Run("rejects duplicate handle", func(t *testing.T) {
		// Add another user with different handle
		otherUser := &domain.User{
			ID:     "user-other",
			Email:  "other@example.com",
			Handle: "existinghandle",
		}
		userRepo.AddUser(otherUser)

		form := url.Values{}
		form.Set("handle", "existinghandle")

		req := httptest.NewRequest(http.MethodPost, "/dashboard/profile", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.UpdateProfile(rec, req)

		if rec.Code != http.StatusConflict {
			t.Errorf("expected 409 conflict, got %d", rec.Code)
		}
	})

	t.Run("responds with Turbo Stream", func(t *testing.T) {
		form := url.Values{}
		form.Set("title", "Turbo Update")

		req := httptest.NewRequest(http.MethodPost, "/dashboard/profile", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "text/vnd.turbo-stream.html")
		rec := httptest.NewRecorder()

		handler.UpdateProfile(rec, req)

		contentType := rec.Header().Get("Content-Type")
		if !strings.Contains(contentType, "turbo-stream") {
			t.Errorf("expected turbo-stream content type, got %s", contentType)
		}
	})
}

func TestUserHandler_UpdateSEO(t *testing.T) {
	userRepo := mocks.NewMockUserRepository()
	sessionMgr := mocks.NewMockSessionManager()
	handler := adapters_http.NewUserHandler(userRepo, sessionMgr)

	user := &domain.User{
		ID:     "user-123",
		Email:  "test@example.com",
		Handle: "testuser",
	}
	userRepo.AddUser(user)
	sessionMgr.SetCurrentUser("user-123")

	t.Run("updates SEO fields", func(t *testing.T) {
		form := url.Values{}
		form.Set("seo_title", "SEO Title")
		form.Set("seo_description", "SEO Description")
		form.Set("seo_image", "https://example.com/og.png")

		req := httptest.NewRequest(http.MethodPost, "/dashboard/seo", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.UpdateSEO(rec, req)

		if rec.Code != http.StatusSeeOther && rec.Code != http.StatusOK {
			t.Errorf("expected redirect or OK, got %d", rec.Code)
		}

		updated, _ := userRepo.GetByID(context.Background(), "user-123")
		if updated.SEOMeta.Title != "SEO Title" {
			t.Errorf("expected SEO title 'SEO Title', got '%s'", updated.SEOMeta.Title)
		}
	})
}

func TestUserHandler_UpdateTheme(t *testing.T) {
	userRepo := mocks.NewMockUserRepository()
	sessionMgr := mocks.NewMockSessionManager()
	handler := adapters_http.NewUserHandler(userRepo, sessionMgr)

	user := &domain.User{
		ID:     "user-123",
		Email:  "test@example.com",
		Handle: "testuser",
	}
	userRepo.AddUser(user)
	sessionMgr.SetCurrentUser("user-123")

	t.Run("updates theme settings", func(t *testing.T) {
		form := url.Values{}
		form.Set("layout", "grid")
		form.Set("primary_color", "#22C55E")
		form.Set("font", "Inter")
		form.Set("fade_in_animation", "on")

		req := httptest.NewRequest(http.MethodPost, "/dashboard/theme", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.UpdateTheme(rec, req)

		if rec.Code != http.StatusSeeOther && rec.Code != http.StatusOK {
			t.Errorf("expected redirect or OK, got %d", rec.Code)
		}

		updated, _ := userRepo.GetByID(context.Background(), "user-123")
		if updated.Theme.LayoutStyle != "grid" {
			t.Errorf("expected layout 'grid', got '%s'", updated.Theme.LayoutStyle)
		}
		if updated.Theme.PrimaryColor != "#22C55E" {
			t.Errorf("expected color '#22C55E', got '%s'", updated.Theme.PrimaryColor)
		}
		if !updated.Theme.FadeInAnimationEnabled {
			t.Error("expected fade_in_animation to be enabled")
		}
	})

	t.Run("disables animations when not checked", func(t *testing.T) {
		// First enable
		user.Theme.FadeInAnimationEnabled = true
		user.Theme.LogoAnimationEnabled = true
		userRepo.Save(context.Background(), user)

		form := url.Values{}
		form.Set("layout", "stacked")
		// Animations NOT included = unchecked

		req := httptest.NewRequest(http.MethodPost, "/dashboard/theme", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.UpdateTheme(rec, req)

		updated, _ := userRepo.GetByID(context.Background(), "user-123")
		if updated.Theme.FadeInAnimationEnabled {
			t.Error("expected fade_in_animation to be disabled")
		}
		if updated.Theme.LogoAnimationEnabled {
			t.Error("expected logo_animation to be disabled")
		}
	})
}
