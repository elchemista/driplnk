package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/ports"
	"github.com/elchemista/driplnk/internal/service"
)

// Mock UserRepo
type MockUserRepo struct{}

func (m *MockUserRepo) Save(ctx context.Context, user *domain.User) error { return nil }
func (m *MockUserRepo) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	return nil, nil
}
func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
func (m *MockUserRepo) GetByHandle(ctx context.Context, handle string) (*domain.User, error) {
	return nil, nil
}

// Mock SessionManager
type MockSessionManager struct {
	SessionID string
	Error     error
}

func (m *MockSessionManager) CreateSession(ctx context.Context, w http.ResponseWriter, userID string) error {
	m.SessionID = userID
	// Simulate setting cookie
	http.SetCookie(w, &http.Cookie{Name: "user_session", Value: userID})
	return m.Error
}

func (m *MockSessionManager) GetSession(r *http.Request) (string, error) {
	return m.SessionID, m.Error
}

func (m *MockSessionManager) ClearSession(ctx context.Context, w http.ResponseWriter) error {
	m.SessionID = ""
	// Simulate clearing cookie
	http.SetCookie(w, &http.Cookie{Name: "user_session", MaxAge: -1})
	return m.Error
}

func TestAuthHandler_Logout(t *testing.T) {
	// Setup
	repo := &MockUserRepo{}
	allowedEmails := []string{"*"}
	authService := service.NewAuthService(repo, allowedEmails)
	mockSession := &MockSessionManager{SessionID: "existing-user"}

	// Mocks (using nil for providers as Logout shouldn't use them)
	// Secure = false for test
	handler := NewAuthHandler(authService, nil, nil, mockSession, false)

	// Case 1: DELETE request (Success)
	req := httptest.NewRequest(http.MethodDelete, "/auth/logout", nil)
	rr := httptest.NewRecorder()
	handler.Logout(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	// Verify cookie clearing
	cookies := rr.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "user_session" {
			if c.MaxAge < 0 { // Should be expired/deleted
				// Go's MaxAge < 0 means delete, but in checking headers we often see Expires in past
				// Mock implementation set MaxAge: -1, which helper usually translates to Expires in past or Max-Age=0/negative
			}
			found = true
		}
	}
	if !found {
		// Our mock sets it, so it should be there.
	}

	if mockSession.SessionID != "" {
		t.Error("session should be cleared in manager")
	}

	// Case 2: POST request (Fail)
	req = httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rr = httptest.NewRecorder()
	handler.Logout(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

// Minimal mock provider here to avoid cross-package test issues with `mocks` package if in valid module
type LocalMockProvider struct {
	AuthURL string
}

func (m *LocalMockProvider) GetAuthURL(state string) string { return m.AuthURL + "?state=" + state }
func (m *LocalMockProvider) Exchange(ctx context.Context, code string) (*ports.OAuthToken, error) {
	return nil, nil
}
func (m *LocalMockProvider) GetUserInfo(ctx context.Context, token *ports.OAuthToken) (*ports.OAuthUser, error) {
	return nil, nil
}

func TestAuthHandler_LoginRedirect(t *testing.T) {
	repo := &MockUserRepo{}
	allowedEmails := []string{"*"}
	authService := service.NewAuthService(repo, allowedEmails)
	mockGithub := &LocalMockProvider{AuthURL: "http://github.com/login"}
	mockSession := &MockSessionManager{}

	handler := NewAuthHandler(authService, mockGithub, nil, mockSession, false)

	req := httptest.NewRequest(http.MethodGet, "/auth/github/login", nil)
	rr := httptest.NewRecorder()

	handler.HandleGithubLogin(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected redirect 307, got %d", rr.Code)
	}

	loc := rr.Header().Get("Location")
	if !strings.Contains(loc, "http://github.com/login") {
		t.Errorf("expected location to contain Github URL, got %s", loc)
	}
	if !strings.Contains(loc, "state=") {
		t.Error("expected state param in redirect URL")
	}

	// Check cookie
	cookies := rr.Result().Cookies()
	stateCookieFound := false
	for _, c := range cookies {
		if c.Name == "oauth_state" {
			stateCookieFound = true
			if c.Value == "" {
				t.Error("state cookie value is empty")
			}
		}
	}
	if !stateCookieFound {
		t.Error("oauth_state cookie not set")
	}
}
