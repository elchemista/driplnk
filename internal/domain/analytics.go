package domain

import (
	"context"
	"time"
)

type AnalyticsEventType string

const (
	EventTypeView   AnalyticsEventType = "view"
	EventTypeClick  AnalyticsEventType = "click"
	EventTypeScroll AnalyticsEventType = "scroll"
)

type AnalyticsEvent struct {
	ID        string             `json:"id"`
	EventType AnalyticsEventType `json:"event_type"`
	LinkID    *string            `json:"link_id,omitempty"` // Optional: for link clicks
	UserID    *string            `json:"user_id,omitempty"` // Optional: owner of the page/link
	VisitorID string             `json:"visitor_id"`
	Country   string             `json:"country,omitempty"`
	Region    string             `json:"region,omitempty"`
	Meta      map[string]string  `json:"meta,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
}

// AnalyticsSummary represents aggregated data for charts/stats
type AnalyticsSummary struct {
	TotalViews  int64            `json:"total_views"`
	TotalClicks int64            `json:"total_clicks"`
	ByCountry   map[string]int64 `json:"by_country"` // country_code -> count
	ByDevice    map[string]int64 `json:"by_device"`  // mobile/desktop -> count (parsed from UA in meta)
}

// AnalyticsRepository defines the contract for analytics data persistence.
// This interface allows swapping between Postgres, Pebble, or other storage backends.
type AnalyticsRepository interface {
	// SaveEvent persists a single analytics event.
	SaveEvent(ctx context.Context, event *AnalyticsEvent) error

	// GetSummary returns aggregated stats for a user (and optionally a specific link).
	// If linkID is empty, it returns stats for the user's profile view.
	GetSummary(ctx context.Context, userID string, linkID *string) (*AnalyticsSummary, error)
}
