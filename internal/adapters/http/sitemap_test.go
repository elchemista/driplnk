package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/elchemista/driplnk/internal/domain"
)

// MockUserRepoForSitemap is a test implementation of domain.UserRepository
type MockUserRepoForSitemap struct {
	users []*domain.User
	err   error
}

func (m *MockUserRepoForSitemap) Save(ctx context.Context, user *domain.User) error {
	return nil
}
func (m *MockUserRepoForSitemap) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	return nil, nil
}
func (m *MockUserRepoForSitemap) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
func (m *MockUserRepoForSitemap) GetByHandle(ctx context.Context, handle string) (*domain.User, error) {
	return nil, nil
}
func (m *MockUserRepoForSitemap) ListAll(ctx context.Context) ([]*domain.User, error) {
	return m.users, m.err
}

func TestSitemapHandler_ServeHTTP_StaticOnly(t *testing.T) {
	// Test with nil UserRepo - should return only static routes
	handler := NewSitemapHandler("http://localhost:8080", nil)

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/xml" {
		t.Errorf("expected Content-Type application/xml, got %s", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "<urlset") {
		t.Error("expected body to contain <urlset")
	}
	if !strings.Contains(body, "http://localhost:8080/") {
		t.Error("expected body to contain base URL")
	}
	if !strings.Contains(body, "http://localhost:8080/login") {
		t.Error("expected body to contain /login")
	}
}

func TestSitemapHandler_ServeHTTP_WithUsers(t *testing.T) {
	// Create mock users
	now := time.Now()
	mockUsers := []*domain.User{
		{
			ID:        "user1",
			Email:     "alice@example.com",
			Handle:    "alice",
			Title:     "Alice",
			UpdatedAt: now,
		},
		{
			ID:        "user2",
			Email:     "bob@example.com",
			Handle:    "bob",
			Title:     "Bob",
			UpdatedAt: now.Add(-24 * time.Hour),
		},
		{
			ID:        "user3",
			Email:     "charlie@example.com",
			Handle:    "charlie",
			Title:     "Charlie",
			UpdatedAt: now.Add(-48 * time.Hour),
		},
	}

	mockRepo := &MockUserRepoForSitemap{users: mockUsers}
	handler := NewSitemapHandler("https://driplnk.io", mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()

	// Check static routes
	if !strings.Contains(body, "https://driplnk.io/") {
		t.Error("expected body to contain base URL")
	}
	if !strings.Contains(body, "https://driplnk.io/login") {
		t.Error("expected body to contain /login")
	}

	// Check dynamic user routes
	if !strings.Contains(body, "https://driplnk.io/alice") {
		t.Error("expected body to contain /alice profile URL")
	}
	if !strings.Contains(body, "https://driplnk.io/bob") {
		t.Error("expected body to contain /bob profile URL")
	}
	if !strings.Contains(body, "https://driplnk.io/charlie") {
		t.Error("expected body to contain /charlie profile URL")
	}

	// Verify XML structure
	if !strings.Contains(body, "<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\"") {
		t.Error("expected proper XML namespace")
	}

	// Check for lastmod dates
	if !strings.Contains(body, "<lastmod>") {
		t.Error("expected lastmod tags in sitemap")
	}

	// Check for changefreq
	if !strings.Contains(body, "<changefreq>weekly</changefreq>") {
		t.Error("expected weekly changefreq for user profiles")
	}

	// Check for priority
	if !strings.Contains(body, "<priority>0.6</priority>") {
		t.Error("expected 0.6 priority for user profiles")
	}
}

func TestSitemapHandler_ServeHTTP_EmptyUsers(t *testing.T) {
	// Test with empty user list
	mockRepo := &MockUserRepoForSitemap{users: []*domain.User{}}
	handler := NewSitemapHandler("http://localhost:8080", mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()

	// Should still have static routes
	if !strings.Contains(body, "http://localhost:8080/") {
		t.Error("expected body to contain base URL")
	}
	if !strings.Contains(body, "http://localhost:8080/login") {
		t.Error("expected body to contain /login")
	}

	// Count URL entries - should be exactly 2 (home + login)
	urlCount := strings.Count(body, "<url>")
	if urlCount != 2 {
		t.Errorf("expected 2 URL entries for empty users, got %d", urlCount)
	}
}

func TestSitemapHandler_ServeHTTP_UserWithEmptyHandle(t *testing.T) {
	// Users with empty handles should be skipped
	mockUsers := []*domain.User{
		{
			ID:        "user1",
			Email:     "test@example.com",
			Handle:    "", // Empty handle - should be skipped
			UpdatedAt: time.Now(),
		},
		{
			ID:        "user2",
			Email:     "valid@example.com",
			Handle:    "validuser",
			UpdatedAt: time.Now(),
		},
	}

	mockRepo := &MockUserRepoForSitemap{users: mockUsers}
	handler := NewSitemapHandler("http://localhost:8080", mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	body := rr.Body.String()

	// Should have validuser
	if !strings.Contains(body, "http://localhost:8080/validuser") {
		t.Error("expected body to contain /validuser")
	}

	// Count URL entries - home + login + 1 valid user = 3
	urlCount := strings.Count(body, "<url>")
	if urlCount != 3 {
		t.Errorf("expected 3 URL entries, got %d", urlCount)
	}
}

func TestSitemapHandler_ServeHTTP_RepoError(t *testing.T) {
	// Test graceful handling when repo returns an error
	mockRepo := &MockUserRepoForSitemap{
		users: nil,
		err:   domain.ErrNotFound, // Simulate error
	}
	handler := NewSitemapHandler("http://localhost:8080", mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should still return OK with static routes
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()

	// Should still have static routes
	if !strings.Contains(body, "http://localhost:8080/") {
		t.Error("expected body to contain base URL despite repo error")
	}
}
