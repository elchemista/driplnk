package mocks

import (
	"context"
	"sync"

	"github.com/elchemista/driplnk/internal/domain"
)

// MockAnalyticsRepository is a test double for domain.AnalyticsRepository.
type MockAnalyticsRepository struct {
	mu     sync.RWMutex
	events []*domain.AnalyticsEvent

	// Hooks for custom behavior
	SaveEventFunc  func(ctx context.Context, event *domain.AnalyticsEvent) error
	GetSummaryFunc func(ctx context.Context, userID string, linkID *string) (*domain.AnalyticsSummary, error)
}

func NewMockAnalyticsRepository() *MockAnalyticsRepository {
	return &MockAnalyticsRepository{
		events: make([]*domain.AnalyticsEvent, 0),
	}
}

func (m *MockAnalyticsRepository) SaveEvent(ctx context.Context, event *domain.AnalyticsEvent) error {
	if m.SaveEventFunc != nil {
		return m.SaveEventFunc(ctx, event)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	e := *event
	m.events = append(m.events, &e)
	return nil
}

func (m *MockAnalyticsRepository) GetSummary(ctx context.Context, userID string, linkID *string) (*domain.AnalyticsSummary, error) {
	if m.GetSummaryFunc != nil {
		return m.GetSummaryFunc(ctx, userID, linkID)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := &domain.AnalyticsSummary{
		ByCountry: make(map[string]int64),
		ByDevice:  make(map[string]int64),
	}

	for _, event := range m.events {
		if event.UserID != nil && *event.UserID == userID {
			if linkID != nil && event.LinkID != nil && *event.LinkID != *linkID {
				continue
			}

			switch event.EventType {
			case domain.EventTypeView:
				summary.TotalViews++
			case domain.EventTypeClick:
				summary.TotalClicks++
			}

			if event.Country != "" {
				summary.ByCountry[event.Country]++
			}
			if device, ok := event.Meta["device_type"]; ok {
				summary.ByDevice[device]++
			}
		}
	}

	return summary, nil
}

// GetEvents returns all recorded events for assertions.
func (m *MockAnalyticsRepository) GetEvents() []*domain.AnalyticsEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*domain.AnalyticsEvent, len(m.events))
	copy(result, m.events)
	return result
}
