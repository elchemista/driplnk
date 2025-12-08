package mocks

import (
	"context"
	"sort"
	"sync"

	"github.com/elchemista/driplnk/internal/domain"
)

// MockLinkRepository is a test double for domain.LinkRepository.
type MockLinkRepository struct {
	mu    sync.RWMutex
	links map[domain.LinkID]*domain.Link

	// Hooks for custom behavior
	SaveFunc       func(ctx context.Context, link *domain.Link) error
	GetByIDFunc    func(ctx context.Context, id domain.LinkID) (*domain.Link, error)
	ListByUserFunc func(ctx context.Context, userID domain.UserID) ([]*domain.Link, error)
	DeleteFunc     func(ctx context.Context, id domain.LinkID) error
	ReorderFunc    func(ctx context.Context, userID domain.UserID, linkIDs []domain.LinkID) error
}

func NewMockLinkRepository() *MockLinkRepository {
	return &MockLinkRepository{
		links: make(map[domain.LinkID]*domain.Link),
	}
}

func (m *MockLinkRepository) Save(ctx context.Context, link *domain.Link) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, link)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	l := *link
	m.links[l.ID] = &l
	return nil
}

func (m *MockLinkRepository) GetByID(ctx context.Context, id domain.LinkID) (*domain.Link, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	if link, ok := m.links[id]; ok {
		return link, nil
	}
	return nil, domain.ErrNotFound
}

func (m *MockLinkRepository) ListByUser(ctx context.Context, userID domain.UserID) ([]*domain.Link, error) {
	if m.ListByUserFunc != nil {
		return m.ListByUserFunc(ctx, userID)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*domain.Link
	for _, link := range m.links {
		if link.UserID == userID {
			result = append(result, link)
		}
	}

	// Sort by order
	sort.Slice(result, func(i, j int) bool {
		return result[i].Order < result[j].Order
	})

	return result, nil
}

func (m *MockLinkRepository) Delete(ctx context.Context, id domain.LinkID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.links, id)
	return nil
}

func (m *MockLinkRepository) Reorder(ctx context.Context, userID domain.UserID, linkIDs []domain.LinkID) error {
	if m.ReorderFunc != nil {
		return m.ReorderFunc(ctx, userID, linkIDs)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, id := range linkIDs {
		if link, ok := m.links[id]; ok && link.UserID == userID {
			link.Order = i
		}
	}
	return nil
}

// AddLink is a helper to seed links for tests.
func (m *MockLinkRepository) AddLink(link *domain.Link) {
	m.mu.Lock()
	defer m.mu.Unlock()

	l := *link
	m.links[l.ID] = &l
}
