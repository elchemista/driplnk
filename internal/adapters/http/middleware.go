package http

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/service"
)

type AnalyticsMiddleware struct {
	service *service.AnalyticsService
}

func NewAnalyticsMiddleware(s *service.AnalyticsService) *AnalyticsMiddleware {
	return &AnalyticsMiddleware{service: s}
}

// TrackView is a middleware that records a page view event asynchronously.
// It tries to identify the 'owner' of the page being viewed.
// For now, this is somewhat manual: we need to know WHICH user profile is being viewed.
// Simple approach: The handler implementing the profile view should probably put the userID in context,
// OR this middleware is applied specifically to routes where we can derive userID from URL (e.g. /u/{handle}).
//
// However, standard Middleware runs BEFORE the handler.
// If we use Chi/Mux, we might have params in context.
// Since we use standard `http.ServeMux` (Go 1.22+), we can use `r.PathValue("handle")`.
// BUT, we need to map handle -> userID.
//
// Easier approach for MVP:
// The Middleware wraps the specific handler which already knows the user.
// actually, standard pattern:
// mux.HandleFunc("/MyPage", middleware.Track(MyHandler))
//
// In this case, extracting the "Subject User ID" (the profile owner) from a generic middleware is hard without querying DB.
// AND we want to avoid DB query in middleware if possible, or at least share it.
//
// Let's implement a wrapper that takes a UserID extractor or assumes it can find it.
//
// Alternative: We only track "Raw Views" here with URL path, and do resolution later?
// No, we want valid analytics.
//
// Compromise:
// `TrackView` takes a function that extracts the "TargetUserID" from the request.
// Or effectively, we can just log the URL path and resolve it? No, better to be explicit.
//
// Let's stick to the Plan: "Checks if request is for a trackable page"
// I will implement a general `TrackView` that accepts a `TargetUserID` string (optional).
// Actually, `TrackView` might be best used explicitly inside the handler for maximum Context awareness?
// "Middleware" usually implies automatic.
//
// Let's go with: Middleware that tracks the generic request info,
// and if we can't determine UserID easily, we just log the Path.
// But for analytics aggregation we usually need UserID.
//
// For simplicity in this Hotwire/Server Refactor:
// I'll create `TrackView(next http.Handler) http.Handler`
// It will launch the async goroutine.
// It will try to guess UserID from Context if set by previous middleware (like "LoadUserFromHandle"),
// or it just records path and we can solve "who owns this path" at ingestion or later.
//
// WAIT: The `AnalyticsEvent` has `UserID *string` (owner).
// If we don't satisfy this, aggregation `WHERE user_id = ...` fails.
//
// Strategy:
// We'll trust the Context to contain `CtxKeyTargetUserID` if available.
// If not, we might miss the user association.
//
// BETTER:
// `AnalyticsMiddleware` just provides `ServeHTTP`.
// But we need a way to inject it.
//
// Let's implement `TrackProfileView` specifically which expects a specific context or header.
// OR, since this is "Application Specific", we can make `TrackView` accept a `targetUserResolver` func.

func (m *AnalyticsMiddleware) TrackView(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// execute original handler first?
		// If we do, we might know response status (to avoid tracking 404s).
		// But `http.ResponseWriter` is write-only unless wrapped.

		// Let's wrap standard ResponseWriter to capture status code
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(ww, r)

		if ww.status >= 400 {
			// Don't track errors/404s
			return
		}

		// Async Tracking
		ctx := context.WithoutCancel(r.Context()) // Ensure context persists
		go func() {
			// Extract details
			meta := make(map[string]string)
			meta["path"] = r.URL.Path

			// User Agent & IP
			ua := r.Header.Get("User-Agent")
			if ua != "" {
				meta["user_agent"] = ua
				if strings.Contains(strings.ToLower(ua), "mobile") || strings.Contains(strings.ToLower(ua), "android") {
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

			// Attempt to find Target User ID from Context
			// We define a context key for this.
			var targetUserID *string
			if val := r.Context().Value(domain.CtxKeyTargetUserID); val != nil {
				if uid, ok := val.(string); ok {
					targetUserID = &uid
				}
			}

			// Visitor ID (Cookie)
			// For now, simple reading of "visitor_id" cookie
			visitorID := "unknown"
			if c, err := r.Cookie("drip_visitor"); err == nil {
				visitorID = c.Value
			} else {
				// We rely on client (JS) to set it, OR we set it here.
				// Since we are "Server Side", we should handle Visitor ID.
				// We can't set cookie in goroutine (headers already sent).
				// Logic should be in main thread if we want to set cookie.
				// But let's assume Visitor ID exists or we just accept 'unknown' for now.
			}

			if err := m.service.TrackEvent(ctx, domain.EventTypeView, targetUserID, nil, visitorID, meta); err != nil {
				log.Printf("[ERR] Failed to track view: %v", err)
			}
		}()
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
