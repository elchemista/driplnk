package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/ports"
	"github.com/elchemista/driplnk/internal/service"
	"github.com/elchemista/driplnk/views/dashboard"
)

// containsMobile checks if a User-Agent string indicates a mobile device.
func containsMobile(ua string) bool {
	lower := strings.ToLower(ua)
	return strings.Contains(lower, "mobile") || strings.Contains(lower, "android")
}

type LinkHandler struct {
	linkSvc      *service.LinkService
	analyticsSvc *service.AnalyticsService
	sessions     ports.SessionManager
	userRepo     domain.UserRepository
}

func NewLinkHandler(
	linkSvc *service.LinkService,
	analyticsSvc *service.AnalyticsService,
	sessions ports.SessionManager,
	userRepo domain.UserRepository,
) *LinkHandler {
	return &LinkHandler{
		linkSvc:      linkSvc,
		analyticsSvc: analyticsSvc,
		sessions:     sessions,
		userRepo:     userRepo,
	}
}

// getCurrentUser retrieves the authenticated user from session.
func (h *LinkHandler) getCurrentUser(r *http.Request) (*domain.User, error) {
	sessionUserID, err := h.sessions.GetSession(r)
	if err != nil || sessionUserID == "" {
		return nil, fmt.Errorf("no session")
	}
	return h.userRepo.GetByID(r.Context(), domain.UserID(sessionUserID))
}

// HandleRedirect handles the /go/{id} endpoint.
// It tracks the click and redirects to the target URL.
func (h *LinkHandler) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	linkIDStr := r.PathValue("id")
	if linkIDStr == "" {
		http.NotFound(w, r)
		return
	}
	linkID := domain.LinkID(linkIDStr)

	ctx := r.Context()
	link, err := h.linkSvc.GetLink(ctx, linkID)
	if err != nil || link == nil {
		http.NotFound(w, r)
		return
	}

	if !link.IsActive {
		http.NotFound(w, r)
		return
	}

	// Async tracking
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

	http.Redirect(w, r, link.URL, http.StatusTemporaryRedirect)
}

// CreateLink handles POST /dashboard/links
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.getCurrentUser(r)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	url := strings.TrimSpace(r.FormValue("url"))
	linkType := domain.LinkType(r.FormValue("type"))

	if title == "" || url == "" {
		respondError(w, r, "Title and URL are required", http.StatusBadRequest)
		return
	}

	link, err := h.linkSvc.CreateLink(r.Context(), user.ID, title, url, linkType)
	if err != nil {
		log.Printf("[ERR] Failed to create link: %v", err)
		respondError(w, r, "Failed to create link", http.StatusInternalServerError)
		return
	}

	// Respond with Turbo Stream or redirect
	if IsTurboRequest(r) {
		w.Header().Set("Content-Type", "text/vnd.turbo-stream.html; charset=utf-8")
		// Use the template-based stream response for proper social link rendering
		dashboard.LinkItemStream(link).Render(r.Context(), w)
		return
	}

	TurboAwareRedirect(w, r, "/dashboard?tab=links")
}

// UpdateLink handles POST /dashboard/links/{id}
func (h *LinkHandler) UpdateLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.getCurrentUser(r)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	linkID := domain.LinkID(r.PathValue("id"))
	if linkID == "" {
		http.Error(w, "Link ID required", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	var title, url *string
	var linkType *domain.LinkType
	var isActive *bool

	if t := r.FormValue("title"); t != "" {
		title = &t
	}
	if u := r.FormValue("url"); u != "" {
		url = &u
	}
	if lt := r.FormValue("type"); lt != "" {
		ltt := domain.LinkType(lt)
		linkType = &ltt
	}
	if a := r.FormValue("is_active"); a != "" {
		active := a == "true" || a == "1" || a == "on"
		isActive = &active
	}

	link, err := h.linkSvc.UpdateLink(r.Context(), linkID, user.ID, title, url, linkType, isActive)
	if err != nil {
		log.Printf("[ERR] Failed to update link: %v", err)
		respondError(w, r, "Failed to update link", http.StatusInternalServerError)
		return
	}

	if IsTurboRequest(r) {
		w.Header().Set("Content-Type", "text/vnd.turbo-stream.html; charset=utf-8")
		fmt.Fprintf(w, `<turbo-stream action="replace" target="link-%s">
  <template>
    <div id="link-%s" class="card border border-base-300 bg-base-100 shadow-sm">
      <div class="card-body gap-2">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div>
            <p class="text-lg font-semibold">%s</p>
            <p class="text-sm text-base-content/70">%s</p>
          </div>
          <div class="flex items-center gap-2">
            <button type="button" class="btn btn-ghost btn-sm" data-action="click->link#edit" data-link-id="%s">Edit</button>
            <form method="post" action="/dashboard/links/%s/delete" data-turbo-confirm="Are you sure?">
              <button type="submit" class="btn btn-ghost btn-sm text-error">Delete</button>
            </form>
          </div>
        </div>
      </div>
    </div>
  </template>
</turbo-stream>`, link.ID, link.ID, link.Title, link.URL, link.ID, link.ID)
		return
	}

	TurboAwareRedirect(w, r, "/dashboard?tab=links")
}

// DeleteLink handles POST /dashboard/links/{id}/delete
func (h *LinkHandler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.getCurrentUser(r)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	linkID := domain.LinkID(r.PathValue("id"))
	if linkID == "" {
		http.Error(w, "Link ID required", http.StatusBadRequest)
		return
	}

	if err := h.linkSvc.DeleteLink(r.Context(), linkID, user.ID); err != nil {
		log.Printf("[ERR] Failed to delete link: %v", err)
		respondError(w, r, "Failed to delete link", http.StatusInternalServerError)
		return
	}

	if IsTurboRequest(r) {
		w.Header().Set("Content-Type", "text/vnd.turbo-stream.html; charset=utf-8")
		fmt.Fprintf(w, `<turbo-stream action="remove" target="link-%s"></turbo-stream>`, linkID)
		return
	}

	TurboAwareRedirect(w, r, "/dashboard?tab=links")
}

// RefreshLinkMetadata handles POST /dashboard/links/{id}/refresh
func (h *LinkHandler) RefreshLinkMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.getCurrentUser(r)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	linkID := domain.LinkID(r.PathValue("id"))
	if linkID == "" {
		http.Error(w, "Link ID required", http.StatusBadRequest)
		return
	}

	link, err := h.linkSvc.RefreshMetadata(r.Context(), linkID, user.ID)
	if err != nil {
		log.Printf("[ERR] Failed to refresh metadata: %v", err)
		respondError(w, r, "Failed to refresh metadata", http.StatusInternalServerError)
		return
	}

	if IsTurboRequest(r) {
		w.Header().Set("Content-Type", "text/vnd.turbo-stream.html; charset=utf-8")
		fmt.Fprintf(w, `<turbo-stream action="append" target="flash-messages">
  <template>
    <div class="alert alert-success shadow-lg mb-4" data-controller="flash">
      <span>Metadata refreshed for %s!</span>
    </div>
  </template>
</turbo-stream>`, link.Title)
		return
	}

	TurboAwareRedirect(w, r, "/dashboard?tab=links")
}

// ReorderLinks handles POST /dashboard/links/reorder
func (h *LinkHandler) ReorderLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.getCurrentUser(r)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var payload struct {
		OrderedIDs []string `json:"ordered_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	linkIDs := make([]domain.LinkID, len(payload.OrderedIDs))
	for i, id := range payload.OrderedIDs {
		linkIDs[i] = domain.LinkID(id)
	}

	if err := h.linkSvc.ReorderLinks(r.Context(), user.ID, linkIDs); err != nil {
		log.Printf("[ERR] Failed to reorder links: %v", err)
		http.Error(w, "Failed to reorder", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// respondError sends an error response appropriate for the request type.
func respondError(w http.ResponseWriter, r *http.Request, message string, status int) {
	if IsTurboRequest(r) {
		w.Header().Set("Content-Type", "text/vnd.turbo-stream.html; charset=utf-8")
		w.WriteHeader(status)
		fmt.Fprintf(w, `<turbo-stream action="append" target="flash-messages">
  <template>
    <div class="alert alert-error shadow-lg mb-4" data-controller="flash">
      <span>%s</span>
    </div>
  </template>
</turbo-stream>`, message)
		return
	}
	http.Error(w, message, status)
}
