// Package mocks provides mock implementations of domain interfaces for testing.
package mocks

import (
	"context"
	"sync"

	"github.com/elchemista/driplnk/internal/domain"
)

// MockUserRepository is a test double for domain.UserRepository.
type MockUserRepository struct {
	mu       sync.RWMutex
	users    map[domain.UserID]*domain.User
	byEmail  map[string]*domain.User
	byHandle map[string]*domain.User

	// Hooks for custom behavior
	SaveFunc        func(ctx context.Context, user *domain.User) error
	GetByIDFunc     func(ctx context.Context, id domain.UserID) (*domain.User, error)
	GetByEmailFunc  func(ctx context.Context, email string) (*domain.User, error)
	GetByHandleFunc func(ctx context.Context, handle string) (*domain.User, error)
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:    make(map[domain.UserID]*domain.User),
		byEmail:  make(map[string]*domain.User),
		byHandle: make(map[string]*domain.User),
	}
}

func (m *MockUserRepository) Save(ctx context.Context, user *domain.User) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, user)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Copy user to avoid shared state issues
	u := *user
	m.users[u.ID] = &u
	m.byEmail[u.Email] = &u
	m.byHandle[u.Handle] = &u
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, domain.ErrNotFound
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	if user, ok := m.byEmail[email]; ok {
		return user, nil
	}
	return nil, domain.ErrNotFound
}

func (m *MockUserRepository) GetByHandle(ctx context.Context, handle string) (*domain.User, error) {
	if m.GetByHandleFunc != nil {
		return m.GetByHandleFunc(ctx, handle)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	if user, ok := m.byHandle[handle]; ok {
		return user, nil
	}
	return nil, domain.ErrNotFound
}

// AddUser is a helper to seed users for tests.
func (m *MockUserRepository) AddUser(user *domain.User) {
	m.mu.Lock()
	defer m.mu.Unlock()

	u := *user
	m.users[u.ID] = &u
	m.byEmail[u.Email] = &u
	m.byHandle[u.Handle] = &u
}
