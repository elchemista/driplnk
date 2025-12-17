package http

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/pkg/sanitizer"
	"github.com/elchemista/driplnk/internal/pkg/validator"
	"github.com/elchemista/driplnk/internal/ports"
	"github.com/elchemista/driplnk/views/dashboard"
)

// UserHandler handles profile, SEO, and theme updates.
type UserHandler struct {
	users    domain.UserRepository
	sessions ports.SessionManager
	uploader ports.FileUploader
}

func NewUserHandler(users domain.UserRepository, sessions ports.SessionManager, uploader ports.FileUploader) *UserHandler {
	return &UserHandler{
		users:    users,
		sessions: sessions,
		uploader: uploader,
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

	// Parse form - support both multipart (for file uploads) and regular forms
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}
	}

	// Extract and sanitize form values
	title := sanitizer.Normalize(r.FormValue("title"))
	handle := slugifyHandle(sanitizer.Normalize(r.FormValue("handle"))) // Apply slugify here
	description := sanitizer.Normalize(r.FormValue("description"))

	// Handle Avatar Upload
	avatarURL := user.AvatarURL // Default to existing
	newAvatarUploaded := false
	file, header, err := r.FormFile("avatar")
	if err == nil {
		defer file.Close()
		// If a file was uploaded and uploader is configured, process it
		if h.uploader != nil {
			// Generate filename: userid_timestamp.ext
			ext := ".jpg"
			if strings.HasSuffix(strings.ToLower(header.Filename), ".png") {
				ext = ".png"
			}
			filename := fmt.Sprintf("%s_%d%s", user.ID, time.Now().Unix(), ext)

			uploadedURL, err := h.uploader.UploadProfileImage(r.Context(), file, filename)
			if err != nil {
				log.Printf("[ERR] Failed to upload avatar: %v", err)
			} else {
				avatarURL = uploadedURL
				newAvatarUploaded = true
			}
		}
	} else if err != http.ErrMissingFile {
		log.Printf("[WARN] Error retrieving avatar file: %v", err)
	}

	// Check uniqueness if handle changed and is not empty
	if handle != "" && handle != user.Handle {
		existing, _ := h.users.GetByHandle(r.Context(), handle)
		if existing != nil && existing.ID != user.ID {
			respondError(w, r, "Handle already taken", http.StatusConflict)
			return
		}
	}

	// Update user fields
	user.Title = title
	user.Handle = handle
	user.Description = description
	user.AvatarURL = avatarURL
	// Also update SEO image if a new avatar was uploaded
	if newAvatarUploaded {
		user.SEOMeta.ImageURL = avatarURL
	}
	user.UpdatedAt = time.Now()

	if err := validator.ValidateStruct(user); err != nil {
		respondError(w, r, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.users.Save(r.Context(), user); err != nil {
		log.Printf("[ERR] Failed to update profile: %v", err)
		respondError(w, r, "Failed to save profile", http.StatusInternalServerError)
		return
	}

	// Respond with success
	if IsTurboRequest(r) {
		w.Header().Set("Content-Type", "text/vnd.turbo-stream.html; charset=utf-8")
		// 1. Flash message
		// 2. We can also reload the frame by redirecting via Turbo script?
		// Or just append the flash and let the client-side value persist (it's already correct due to js).
		fmt.Fprintf(w, `<turbo-stream action="append" target="flash-messages">
  <template>
    <div class="alert alert-success shadow-lg mb-4" data-controller="flash">
      <span>Profile updated successfully!</span>
    </div>
  </template>
</turbo-stream>`)
		return
	}

	TurboAwareRedirect(w, r, "/dashboard?tab=profile")
}

func slugifyHandle(s string) string {
	// Slugify: preserve case, remove spaces, remove non-alphanumeric (except - and _)
	s = strings.TrimSpace(s)
	// Remove spaces completely
	s = strings.ReplaceAll(s, " ", "")
	// Remove invalid chars (keep only a-zA-Z0-9-_)
	reg, err := regexp.Compile("[^a-zA-Z0-9-_]+")
	if err != nil {
		return s
	}
	return reg.ReplaceAllString(s, "")
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

	// Update SEO fields - allow empty values to clear
	if r.Form.Has("seo_title") {
		user.SEOMeta.Title = sanitizer.Normalize(r.FormValue("seo_title"))
	}
	if r.Form.Has("seo_description") {
		user.SEOMeta.Description = sanitizer.Normalize(r.FormValue("seo_description"))
	}
	if r.Form.Has("seo_image") {
		user.SEOMeta.ImageURL = sanitizer.SanitizeURL(r.FormValue("seo_image"))
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
    <div class="alert alert-success shadow-lg mb-4" data-controller="flash">
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

	// Parse mode
	if mode := r.FormValue("mode"); mode != "" {
		user.Theme.Mode = mode
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
		// 1. Flash message
		fmt.Fprintf(w, `<turbo-stream action="append" target="flash-messages">
  <template>
    <div class="alert alert-success shadow-lg mb-4" data-controller="flash">
      <span>Theme updated!</span>
    </div>
  </template>
</turbo-stream>`)

		// 2. Preview Update
		if err := dashboard.ThemePreviewStream(user).Render(r.Context(), w); err != nil {
			log.Printf("[ERR] Failed to render theme preview: %v", err)
		}
		return
	}

	TurboAwareRedirect(w, r, "/dashboard?tab=theme")
}
