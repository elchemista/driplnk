package http

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/service"
)

// containsMobile checks if a User-Agent string indicates a mobile device.
func containsMobile(ua string) bool {
	lower := strings.ToLower(ua)
	return strings.Contains(lower, "mobile") || strings.Contains(lower, "android")
}

type LinkHandler struct {
	linkRepo     domain.LinkRepository
	analyticsSvc *service.AnalyticsService
}

func NewLinkHandler(linkRepo domain.LinkRepository, analyticsSvc *service.AnalyticsService) *LinkHandler {
	return &LinkHandler{
		linkRepo:     linkRepo,
		analyticsSvc: analyticsSvc,
	}
}

// HandleRedirect handles the /go/{id} endpoint.
// It tracks the click and redirects to the target URL.
func (h *LinkHandler) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	// 1. Extract Link ID from path
	// Assuming standard ServeMux in Go 1.22+
	linkIDStr := r.PathValue("id")
	if linkIDStr == "" {
		http.NotFound(w, r)
		return
	}
	linkID := domain.LinkID(linkIDStr)

	// 2. Fetch Link to get URL and Owner
	ctx := r.Context()
	link, err := h.linkRepo.GetByID(ctx, linkID)
	if err != nil {
		// Link not found or error
		http.NotFound(w, r)
		return
	}

	// 3. Track Click (Async)
	// We want to track it even if the link is "inactive"?
	// Probably not if we don't redirect.
	// But let's check active status first.
	if !link.IsActive {
		// Show 404 or specific "Link Disabled" page
		http.NotFound(w, r)
		return
	}

	// Async tracking
	// Create detached context or just background context with values?
	// We need request headers.
	go func() {
		trackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		meta := make(map[string]string)
		meta["path"] = r.URL.Path

		ua := r.Header.Get("User-Agent")
		if ua != "" {
			meta["user_agent"] = ua
			if containsMobile(ua) {
				meta["device_type"] = "mobile"
			} else {
				meta["device_type"] = "desktop"
			}
		}

		country := r.Header.Get("CF-IPCountry")
		if country == "" {
			country = r.Header.Get("X-AppEngine-Country")
		}
		if country != "" {
			meta["country"] = country
		}

		visitorID := "unknown"
		if c, err := r.Cookie("drip_visitor"); err == nil {
			visitorID = c.Value
		}

		userID := string(link.UserID)
		lID := string(link.ID)

		if err := h.analyticsSvc.TrackEvent(trackCtx, domain.EventTypeClick, &userID, &lID, visitorID, meta); err != nil {
			log.Printf("[ERR] Failed to track click for link %s: %v", link.ID, err)
		}
	}()

	// 4. Redirect
	// 307 Temporary Redirect preserves method, but for links 302/307 is fine.
	// 307 is "Temporary Redirect".
	http.Redirect(w, r, link.URL, http.StatusTemporaryRedirect)
}
