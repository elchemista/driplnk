package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cockroachdb/pebble"
	"github.com/elchemista/driplnk/internal/domain"
)

// SaveEvent persists a single analytics event in PebbleDB.
func (r *PebbleRepository) SaveEvent(ctx context.Context, event *domain.AnalyticsEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	batch := r.db.NewBatch()
	defer batch.Close()

	// 1. Main record: analytics:event:<event_id>
	eventKey := []byte(fmt.Sprintf("analytics:event:%s", event.ID))
	if err := batch.Set(eventKey, data, pebble.Sync); err != nil {
		return err
	}

	// 2. User Index: analytics:user:<user_id>:<event_id>
	// This helps us scan only events for a specific user.
	if event.UserID != nil {
		userIndexKey := []byte(fmt.Sprintf("analytics:user:%s:%s", *event.UserID, event.ID))
		if err := batch.Set(userIndexKey, []byte{}, pebble.Sync); err != nil {
			return err
		}
	}

	// 3. Link Index: analytics:link:<link_id>:<event_id>
	if event.LinkID != nil {
		linkIndexKey := []byte(fmt.Sprintf("analytics:link:%s:%s", *event.LinkID, event.ID))
		if err := batch.Set(linkIndexKey, []byte{}, pebble.Sync); err != nil {
			return err
		}
	}

	return batch.Commit(pebble.Sync)
}

// GetSummary returns aggregated stats for a user (and optionally a specific link).
func (r *PebbleRepository) GetSummary(ctx context.Context, userID string, linkID *string) (*domain.AnalyticsSummary, error) {
	summary := &domain.AnalyticsSummary{
		ByCountry: make(map[string]int64),
		ByDevice:  make(map[string]int64),
	}

	var prefix []byte
	if linkID != nil {
		// If filtering by link, we can scan the link index directly IF we assume the user owns the link.
		// However, the interface asks for (userID, linkID).
		// We'll trust the caller enforces ownership or we just scan the link index.
		// Key: analytics:link:<link_id>:<event_id>
		prefix = []byte(fmt.Sprintf("analytics:link:%s:", *linkID))
	} else {
		// Scan user index
		// Key: analytics:user:<user_id>:<event_id>
		prefix = []byte(fmt.Sprintf("analytics:user:%s:", userID))
	}

	iter, _ := r.db.NewIter(&pebble.IterOptions{
		LowerBound: prefix,
	})
	defer iter.Close()

	for iter.SeekGE(prefix); iter.Valid() && strings.HasPrefix(string(iter.Key()), string(prefix)); iter.Next() {
		// Extract EventID
		keyParts := strings.Split(string(iter.Key()), ":")
		// analytics:user:<user_id>:<event_id> -> len 4
		// analytics:link:<link_id>:<event_id> -> len 4
		if len(keyParts) < 4 {
			continue
		}
		eventID := keyParts[3]

		// Fetch Event
		eventKey := []byte(fmt.Sprintf("analytics:event:%s", eventID))
		val, closer, err := r.db.Get(eventKey)
		if err != nil {
			// If missing (orphaned index?), skip
			continue
		}

		var event domain.AnalyticsEvent
		if err := json.Unmarshal(val, &event); err != nil {
			closer.Close()
			continue
		}
		closer.Close()

		// Filter by UserID if we are scanning via Link Index (to ensure ownership/correctness)
		if linkID != nil && event.UserID != nil && *event.UserID != userID {
			continue
		}

		// Aggregate
		switch event.EventType {
		case domain.EventTypeView:
			summary.TotalViews++
		case domain.EventTypeClick:
			summary.TotalClicks++
		}

		if event.Country != "" {
			summary.ByCountry[event.Country]++
		}

		if deviceType, ok := event.Meta["device_type"]; ok {
			summary.ByDevice[deviceType]++
		}
	}

	return summary, nil
}
