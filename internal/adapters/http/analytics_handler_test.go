package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/service"
)

type mockAnalyticsRepo struct {
	events []*domain.AnalyticsEvent
}

func (m *mockAnalyticsRepo) SaveEvent(ctx context.Context, event *domain.AnalyticsEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockAnalyticsRepo) GetSummary(ctx context.Context, userID string, linkID *string) (*domain.AnalyticsSummary, error) {
	return nil, nil
}

func TestRecordScroll(t *testing.T) {
	repo := &mockAnalyticsRepo{}
	svc := service.NewAnalyticsService(repo)
	handler := NewAnalyticsHandler(svc)

	payload := map[string]interface{}{
		"visitor_id": "visitor-123",
		"depth":      75,
		"meta": map[string]string{
			"page": "/user/profile",
		},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/analytics/scroll", bytes.NewBuffer(body))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Mobile)")
	req.Header.Set("CF-IPCountry", "US")

	rr := httptest.NewRecorder()
	handler.RecordScroll(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusAccepted)
	}

	if len(repo.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(repo.events))
	}

	event := repo.events[0]
	if event.EventType != domain.EventTypeScroll {
		t.Errorf("expected event type scroll, got %v", event.EventType)
	}
	if event.VisitorID != "visitor-123" {
		t.Errorf("expected visitor id visitor-123, got %v", event.VisitorID)
	}
	if event.Country != "US" {
		t.Errorf("expected country US, got %v", event.Country)
	}
	if event.Meta["device_type"] != "mobile" {
		t.Errorf("expected device type mobile, got %v", event.Meta["device_type"])
	}
}
