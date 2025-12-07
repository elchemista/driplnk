package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/service"
)

type AnalyticsHandler struct {
	service *service.AnalyticsService
}

func NewAnalyticsHandler(s *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: s}
}

// RecordScrollRequest is a lightweight request for scroll events.
type RecordScrollRequest struct {
	UserID    *string           `json:"user_id,omitempty"` // Owner of the page
	VisitorID string            `json:"visitor_id,omitempty"`
	Depth     int               `json:"depth"` // Scroll depth percentage (0-100)
	Meta      map[string]string `json:"meta,omitempty"`
}

// RecordScroll handles scroll depth events sent from client-side Stimulus.
// This is the only client-side event that requires an API call.
func (h *AnalyticsHandler) RecordScroll(w http.ResponseWriter, r *http.Request) {
	var req RecordScrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Enrich meta
	if req.Meta == nil {
		req.Meta = make(map[string]string)
	}

	// Add scroll depth to meta
	req.Meta["scroll_depth"] = string(rune(req.Depth)) // simple encoding

	// Extract User-Agent
	ua := r.Header.Get("User-Agent")
	if ua != "" {
		req.Meta["user_agent"] = ua
		if strings.Contains(strings.ToLower(ua), "mobile") || strings.Contains(strings.ToLower(ua), "android") {
			req.Meta["device_type"] = "mobile"
		} else {
			req.Meta["device_type"] = "desktop"
		}
	}

	// Extract Country
	country := r.Header.Get("CF-IPCountry")
	if country == "" {
		country = r.Header.Get("X-AppEngine-Country")
	}
	if country != "" {
		req.Meta["country"] = country
	}

	err := h.service.TrackEvent(r.Context(), domain.EventTypeScroll, req.UserID, nil, req.VisitorID, req.Meta)
	if err != nil {
		http.Error(w, "failed to record scroll event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
