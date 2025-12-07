package ports

import (
	"context"
	"net/http"
)

// SessionManager defines how sessions are created, retrieved, and destroyed.
// It abstracts away the underlying mechanism (cookie, redis, jwt, etc).
type SessionManager interface {
	// CreateSession creates a new session for the given user ID and sets it on the response.
	CreateSession(ctx context.Context, w http.ResponseWriter, userID string) error

	// GetSession retrieves the user ID from the current request/session.
	// Returns empty string and error if no session found or invalid.
	GetSession(r *http.Request) (string, error)

	// ClearSession invalidates/removes the session from the response.
	ClearSession(ctx context.Context, w http.ResponseWriter) error
}
