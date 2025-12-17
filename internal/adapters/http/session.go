package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/elchemista/driplnk/internal/ports"
	"github.com/gorilla/securecookie"
)

type CookieSessionManager struct {
	cookieName string
	secure     bool
	httpOnly   bool
	domain     string
	path       string
	sc         *securecookie.SecureCookie
}

// Ensure CookieSessionManager implements ports.SessionManager
var _ ports.SessionManager = (*CookieSessionManager)(nil)

func NewCookieSessionManager(secure bool, domain, secretKey string) *CookieSessionManager {
	if secretKey == "" {
		// Fallback for dev/test - DO NOT USE IN PROD without warning
		// But for now we just generate a random one which invalidates sessions on restart
		// Ideally we should panic or error if empty in prod
		secretKey = string(securecookie.GenerateRandomKey(64))
	}

	return &CookieSessionManager{
		cookieName: "user_session",
		secure:     secure,
		httpOnly:   true,
		domain:     domain,
		path:       "/",
		sc:         securecookie.New([]byte(secretKey), nil), // Sign only for now, can add encryption later
	}
}

func (m *CookieSessionManager) CreateSession(ctx context.Context, w http.ResponseWriter, userID string) error {
	encoded, err := m.sc.Encode(m.cookieName, userID)
	if err != nil {
		return fmt.Errorf("failed to encode session: %w", err)
	}

	cookie := &http.Cookie{
		Name:     m.cookieName,
		Value:    encoded,
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

	var userID string
	if err := m.sc.Decode(m.cookieName, cookie.Value, &userID); err != nil {
		return "", fmt.Errorf("invalid session: %w", err)
	}

	return userID, nil
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
