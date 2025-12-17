package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	adapter "github.com/elchemista/driplnk/internal/adapters/http"
	"github.com/stretchr/testify/assert"
)

func TestCookieSessionManager(t *testing.T) {
	// Use a fixed secret for testing
	secret := "super-secret-key-32-bytes-long!!"
	manager := adapter.NewCookieSessionManager(false, "localhost", secret)

	t.Run("Create and Get Session", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx := context.Background()
		userID := "user-123"

		// Create Session
		err := manager.CreateSession(ctx, w, userID)
		assert.NoError(t, err)

		// Parse the cookie from response
		resp := w.Result()
		cookies := resp.Cookies()
		assert.Len(t, cookies, 1)
		sessionCookie := cookies[0]
		assert.Equal(t, "user_session", sessionCookie.Name)
		assert.NotEmpty(t, sessionCookie.Value)
		assert.NotEqual(t, userID, sessionCookie.Value, "Value should be encoded/signed")

		// Get Session
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(sessionCookie)

		decodedID, err := manager.GetSession(req)
		assert.NoError(t, err)
		assert.Equal(t, userID, decodedID)
	})

	t.Run("Tampered Cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx := context.Background()
		userID := "user-123"

		manager.CreateSession(ctx, w, userID)
		cookie := w.Result().Cookies()[0]

		// Tamper with the value
		cookie.Value += "tampered"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(cookie)

		_, err := manager.GetSession(req)
		assert.Error(t, err)
	})
}
