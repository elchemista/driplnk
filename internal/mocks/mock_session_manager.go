package mocks

import (
	"context"
	"net/http"
	"sync"
)

// MockSessionManager is a test double for ports.SessionManager.
type MockSessionManager struct {
	mu       sync.RWMutex
	sessions map[string]string // cookie value -> user ID

	// Track method calls
	CreateCalls []string
	GetCalls    int
	ClearCalls  int

	// Hooks for custom behavior
	CreateSessionFunc func(ctx context.Context, w http.ResponseWriter, userID string) error
	GetSessionFunc    func(r *http.Request) (string, error)
	ClearSessionFunc  func(ctx context.Context, w http.ResponseWriter) error

	// Default session for simple tests
	CurrentUserID string
}

func NewMockSessionManager() *MockSessionManager {
	return &MockSessionManager{
		sessions:    make(map[string]string),
		CreateCalls: make([]string, 0),
	}
}

func (m *MockSessionManager) CreateSession(ctx context.Context, w http.ResponseWriter, userID string) error {
	if m.CreateSessionFunc != nil {
		return m.CreateSessionFunc(ctx, w, userID)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CreateCalls = append(m.CreateCalls, userID)
	m.CurrentUserID = userID
	return nil
}

func (m *MockSessionManager) GetSession(r *http.Request) (string, error) {
	if m.GetSessionFunc != nil {
		return m.GetSessionFunc(r)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetCalls++
	return m.CurrentUserID, nil
}

func (m *MockSessionManager) ClearSession(ctx context.Context, w http.ResponseWriter) error {
	if m.ClearSessionFunc != nil {
		return m.ClearSessionFunc(ctx, w)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ClearCalls++
	m.CurrentUserID = ""
	return nil
}

// SetCurrentUser is a helper to set the logged-in user for tests.
func (m *MockSessionManager) SetCurrentUser(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CurrentUserID = userID
}
