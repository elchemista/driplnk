package http

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/ports"
)

// UserHandler handles profile, SEO, and theme updates.
type UserHandler struct {
	users    domain.UserRepository
	sessions ports.SessionManager
}

func NewUserHandler(users domain.UserRepository, sessions ports.SessionManager) *UserHandler {
	return &UserHandler{
		users:    users,
		sessions: sessions,
	}
}

// getCurrentUser retrieves the authenticated user from session.
func (h *UserHandler) getCurrentUser(r *http.Request) (*domain.User, error) {
	sessionUserID, err := h.sessions.GetSession(r)
	if err != nil || sessionUserID == "" {
		return nil, fmt.Errorf("no session")
	}
	return h.users.GetByID(r.Context(), domain.UserID(sessionUserID))
}

// UpdateProfile handles POST /dashboard/profile
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
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

	// Update fields if provided
	if title := strings.TrimSpace(r.FormValue("title")); title != "" {
		user.Title = title
	}
	if handle := strings.TrimSpace(r.FormValue("handle")); handle != "" {
		// Check uniqueness if handle changed
		if handle != user.Handle {
			existing, _ := h.users.GetByHandle(r.Context(), handle)
			if existing != nil && existing.ID != user.ID {
				respondError(w, r, "Handle already taken", http.StatusConflict)
				return
			}
			user.Handle = handle
		}
	}
	if desc := r.FormValue("description"); desc != "" {
		user.Description = desc
	}
	if avatar := strings.TrimSpace(r.FormValue("avatar")); avatar != "" {
		user.AvatarURL = avatar
	}

	user.UpdatedAt = time.Now()

	if err := h.users.Save(r.Context(), user); err != nil {
		log.Printf("[ERR] Failed to update profile: %v", err)
		respondError(w, r, "Failed to save profile", http.StatusInternalServerError)
		return
	}

	// Respond with success
	if IsTurboRequest(r) {
		w.Header().Set("Content-Type", "text/vnd.turbo-stream.html; charset=utf-8")
		fmt.Fprintf(w, `<turbo-stream action="append" target="flash-messages">
  <template>
    <div class="alert alert-success shadow-lg mb-4">
      <span>Profile updated successfully!</span>
    </div>
  </template>
</turbo-stream>`)
		return
	}

	TurboAwareRedirect(w, r, "/dashboard?tab=profile")
}

// UpdateSEO handles POST /dashboard/seo
func (h *UserHandler) UpdateSEO(w http.ResponseWriter, r *http.Request) {
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

	if title := strings.TrimSpace(r.FormValue("seo_title")); title != "" {
		user.SEOMeta.Title = title
	}
	if desc := r.FormValue("seo_description"); desc != "" {
		user.SEOMeta.Description = desc
	}
	if image := strings.TrimSpace(r.FormValue("seo_image")); image != "" {
		user.SEOMeta.ImageURL = image
	}

	user.UpdatedAt = time.Now()

	if err := h.users.Save(r.Context(), user); err != nil {
		log.Printf("[ERR] Failed to update SEO: %v", err)
		respondError(w, r, "Failed to save SEO settings", http.StatusInternalServerError)
		return
	}

	if IsTurboRequest(r) {
		w.Header().Set("Content-Type", "text/vnd.turbo-stream.html; charset=utf-8")
		fmt.Fprintf(w, `<turbo-stream action="append" target="flash-messages">
  <template>
    <div class="alert alert-success shadow-lg mb-4">
      <span>SEO settings updated!</span>
    </div>
  </template>
</turbo-stream>`)
		return
	}

	TurboAwareRedirect(w, r, "/dashboard?tab=profile")
}

// UpdateTheme handles POST /dashboard/theme
func (h *UserHandler) UpdateTheme(w http.ResponseWriter, r *http.Request) {
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

	// Parse layout
	if layout := r.FormValue("layout"); layout != "" {
		user.Theme.LayoutStyle = layout
	}

	// Parse primary color
	if color := r.FormValue("primary_color"); color != "" {
		user.Theme.PrimaryColor = color
	}

	// Parse font
	if font := strings.TrimSpace(r.FormValue("font")); font != "" {
		user.Theme.TitleFontStyle = font
	}

	// Parse background
	if bgStyle := r.FormValue("background_style"); bgStyle != "" {
		user.Theme.BackgroundStyle = bgStyle
	}
	if bgValue := r.FormValue("background_value"); bgValue != "" {
		user.Theme.BackgroundValue = bgValue
	}

	// Parse animations (checkboxes - present = checked)
	user.Theme.FadeInAnimationEnabled = r.FormValue("fade_in_animation") == "on" || r.FormValue("fade_in_animation") == "true"
	user.Theme.LogoAnimationEnabled = r.FormValue("logo_animation") == "on" || r.FormValue("logo_animation") == "true"

	user.UpdatedAt = time.Now()

	if err := h.users.Save(r.Context(), user); err != nil {
		log.Printf("[ERR] Failed to update theme: %v", err)
		respondError(w, r, "Failed to save theme", http.StatusInternalServerError)
		return
	}

	if IsTurboRequest(r) {
		w.Header().Set("Content-Type", "text/vnd.turbo-stream.html; charset=utf-8")
		fmt.Fprintf(w, `<turbo-stream action="append" target="flash-messages">
  <template>
    <div class="alert alert-success shadow-lg mb-4">
      <span>Theme updated!</span>
    </div>
  </template>
</turbo-stream>
<turbo-stream action="replace" target="theme-preview">
  <template>
    <div id="theme-preview" class="rounded-2xl border border-base-300 bg-base-100 p-4">
      <p class="text-lg font-semibold">%s</p>
      <p class="text-sm text-base-content/70">@%s</p>
      <div class="mt-3 text-sm text-base-content/60">
        <p>Layout: %s</p>
        <p>Color: %s</p>
      </div>
    </div>
  </template>
</turbo-stream>`, user.Title, user.Handle, user.Theme.LayoutStyle, user.Theme.PrimaryColor)
		return
	}

	TurboAwareRedirect(w, r, "/dashboard?tab=theme")
}
