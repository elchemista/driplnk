package http

import (
	"context"
	"net/http"
	"time"

	"github.com/elchemista/driplnk/internal/ports"
)

type CookieSessionManager struct {
	cookieName string
	secure     bool
	httpOnly   bool
	domain     string
	path       string
	// In the future we might want a signing key here
}

// Ensure CookieSessionManager implements ports.SessionManager
var _ ports.SessionManager = (*CookieSessionManager)(nil)

func NewCookieSessionManager(secure bool, domain string) *CookieSessionManager {
	return &CookieSessionManager{
		cookieName: "user_session",
		secure:     secure,
		httpOnly:   true,
		domain:     domain,
		path:       "/",
	}
}

func (m *CookieSessionManager) CreateSession(ctx context.Context, w http.ResponseWriter, userID string) error {
	// TODO: Sign the userID with a secret key to prevent tampering
	cookie := &http.Cookie{
		Name:     m.cookieName,
		Value:    userID,
		Path:     m.path,
		Domain:   m.domain,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: m.httpOnly,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
	return nil
}

func (m *CookieSessionManager) GetSession(r *http.Request) (string, error) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return "", err
	}
	// TODO: Validate signature
	return cookie.Value, nil
}

func (m *CookieSessionManager) ClearSession(ctx context.Context, w http.ResponseWriter) error {
	cookie := &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     m.path,
		Domain:   m.domain,
		Expires:  time.Unix(0, 0), // Expire immediately
		HttpOnly: m.httpOnly,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
	return nil
}
