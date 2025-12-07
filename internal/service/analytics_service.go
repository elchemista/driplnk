package service

import (
	"context"
	"time"

	"github.com/elchemista/driplnk/internal/domain"
	"github.com/google/uuid"
)

type AnalyticsService struct {
	repo domain.AnalyticsRepository
}

func NewAnalyticsService(repo domain.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) TrackEvent(ctx context.Context, eventType domain.AnalyticsEventType, userID *string, linkID *string, visitorID string, meta map[string]string) error {
	// Simple validation
	if visitorID == "" {
		// fallback to random UUID if not provided (though handler should handle this)
		visitorID = uuid.New().String()
	}

	event := &domain.AnalyticsEvent{
		ID:        uuid.New().String(),
		EventType: eventType,
		UserID:    userID,
		LinkID:    linkID,
		VisitorID: visitorID,
		Meta:      meta,
		CreatedAt: time.Now(),
	}

	// Extract country/region/device from meta if available
	if c, ok := meta["country"]; ok {
		event.Country = c
	}
	if r, ok := meta["region"]; ok {
		event.Region = r
	}

	return s.repo.SaveEvent(ctx, event)
}

func (s *AnalyticsService) GetSummary(ctx context.Context, userID string, linkID *string) (*domain.AnalyticsSummary, error) {
	return s.repo.GetSummary(ctx, userID, linkID)
}
