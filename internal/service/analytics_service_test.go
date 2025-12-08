package service_test

import (
	"context"
	"testing"

	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/mocks"
	"github.com/elchemista/driplnk/internal/service"
)

func TestAnalyticsService_TrackEvent(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockAnalyticsRepository()
	svc := service.NewAnalyticsService(repo)

	t.Run("tracks view event", func(t *testing.T) {
		userID := "user-123"
		visitorID := "visitor-456"
		meta := map[string]string{"path": "/u/handle", "device_type": "mobile"}

		err := svc.TrackEvent(ctx, domain.EventTypeView, &userID, nil, visitorID, meta)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := repo.GetEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		event := events[0]
		if event.EventType != domain.EventTypeView {
			t.Errorf("expected event type view, got %s", event.EventType)
		}
		if *event.UserID != userID {
			t.Errorf("expected user ID %s, got %s", userID, *event.UserID)
		}
		if event.VisitorID != visitorID {
			t.Errorf("expected visitor ID %s, got %s", visitorID, event.VisitorID)
		}
	})

	t.Run("tracks click event with link ID", func(t *testing.T) {
		userID := "user-123"
		linkID := "link-789"
		visitorID := "visitor-456"
		meta := map[string]string{"path": "/go/link-789"}

		err := svc.TrackEvent(ctx, domain.EventTypeClick, &userID, &linkID, visitorID, meta)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := repo.GetEvents()
		// Second event after the view event
		if len(events) < 2 {
			t.Fatalf("expected at least 2 events, got %d", len(events))
		}

		event := events[len(events)-1]
		if event.EventType != domain.EventTypeClick {
			t.Errorf("expected event type click, got %s", event.EventType)
		}
		if event.LinkID == nil || *event.LinkID != linkID {
			t.Errorf("expected link ID %s", linkID)
		}
	})

	t.Run("generates visitor ID if empty", func(t *testing.T) {
		userID := "user-999"
		meta := map[string]string{}

		err := svc.TrackEvent(ctx, domain.EventTypeView, &userID, nil, "", meta)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := repo.GetEvents()
		event := events[len(events)-1]
		if event.VisitorID == "" {
			t.Error("expected visitor ID to be generated")
		}
	})

	t.Run("extracts country from meta", func(t *testing.T) {
		userID := "user-geo"
		visitorID := "visitor-geo"
		meta := map[string]string{"country": "US", "region": "CA"}

		err := svc.TrackEvent(ctx, domain.EventTypeView, &userID, nil, visitorID, meta)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		events := repo.GetEvents()
		event := events[len(events)-1]
		if event.Country != "US" {
			t.Errorf("expected country US, got %s", event.Country)
		}
		if event.Region != "CA" {
			t.Errorf("expected region CA, got %s", event.Region)
		}
	})
}

func TestAnalyticsService_GetSummary(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockAnalyticsRepository()
	svc := service.NewAnalyticsService(repo)
	userID := "user-summary"

	// Track some events
	meta := map[string]string{"device_type": "mobile", "country": "US"}
	svc.TrackEvent(ctx, domain.EventTypeView, &userID, nil, "v1", meta)
	svc.TrackEvent(ctx, domain.EventTypeView, &userID, nil, "v2", meta)
	svc.TrackEvent(ctx, domain.EventTypeClick, &userID, nil, "v1", meta)

	t.Run("returns correct summary", func(t *testing.T) {
		summary, err := svc.GetSummary(ctx, userID, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if summary.TotalViews != 2 {
			t.Errorf("expected 2 views, got %d", summary.TotalViews)
		}
		if summary.TotalClicks != 1 {
			t.Errorf("expected 1 click, got %d", summary.TotalClicks)
		}
		if summary.ByCountry["US"] != 3 {
			t.Errorf("expected 3 events from US, got %d", summary.ByCountry["US"])
		}
		if summary.ByDevice["mobile"] != 3 {
			t.Errorf("expected 3 mobile events, got %d", summary.ByDevice["mobile"])
		}
	})
}
