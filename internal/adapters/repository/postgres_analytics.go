package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elchemista/driplnk/internal/domain"
)

// SaveEvent persists a single analytics event.
func (r *PostgresRepository) SaveEvent(ctx context.Context, event *domain.AnalyticsEvent) error {
	metaBytes, err := json.Marshal(event.Meta)
	if err != nil {
		return fmt.Errorf("marshal meta: %w", err)
	}

	query := `
		INSERT INTO analytics_events (event_type, link_id, user_id, visitor_id, country, region, meta, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.ExecContext(ctx, query,
		event.EventType,
		event.LinkID,
		event.UserID,
		event.VisitorID,
		event.Country,
		event.Region,
		metaBytes,
		event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save analytics event: %w", err)
	}
	return nil
}

// GetSummary returns aggregated stats for a user.
func (r *PostgresRepository) GetSummary(ctx context.Context, userID string, linkID *string) (*domain.AnalyticsSummary, error) {
	summary := &domain.AnalyticsSummary{
		ByCountry: make(map[string]int64),
		ByDevice:  make(map[string]int64),
	}

	// Base filter
	filter := "user_id = $1"
	args := []interface{}{userID}
	if linkID != nil {
		filter += " AND link_id = $2"
		args = append(args, *linkID)
	}

	// 1. Counts (Views and Clicks)
	queryCounts := fmt.Sprintf(`
		SELECT 
			COUNT(*) FILTER (WHERE event_type = 'view'),
			COUNT(*) FILTER (WHERE event_type = 'click')
		FROM analytics_events
		WHERE %s
	`, filter)

	err := r.db.QueryRowContext(ctx, queryCounts, args...).Scan(&summary.TotalViews, &summary.TotalClicks)
	if err != nil {
		return nil, fmt.Errorf("failed to get counts: %w", err)
	}

	// 2. Group by Country
	queryCountry := fmt.Sprintf(`
		SELECT country, COUNT(*)
		FROM analytics_events
		WHERE %s AND country != ''
		GROUP BY country
	`, filter)

	rows, err := r.db.QueryContext(ctx, queryCountry, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get country stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var country string
		var count int64
		if err := rows.Scan(&country, &count); err != nil {
			log.Printf("[WARN] Failed to scan country row: %v", err)
			continue
		}
		summary.ByCountry[country] = count
	}

	// 3. Group by Device (assuming meta->>'device' exists or similar)
	// For now, let's try to extract from meta if possible, or skip if not standardized.
	// As we haven't standardized 'device' in ingestion yet, we can skip or use a placeholder.
	// Let's query by meta->>'device_type' if we decide to store it there.
	queryDevice := fmt.Sprintf(`
		SELECT meta->>'device_type', COUNT(*)
		FROM analytics_events
		WHERE %s AND meta->>'device_type' IS NOT NULL
		GROUP BY meta->>'device_type'
	`, filter)

	rowsDevice, err := r.db.QueryContext(ctx, queryDevice, args...)
	if err == nil {
		defer rowsDevice.Close()
		for rowsDevice.Next() {
			var device string
			var count int64
			if err := rowsDevice.Scan(&device, &count); err == nil {
				summary.ByDevice[device] = count
			}
		}
	} else {
		// Just log, don't fail, maybe column/key doesn't exist
		log.Printf("[DEBUG] Failed to get device stats (expected if no data): %v", err)
	}

	return summary, nil
}
