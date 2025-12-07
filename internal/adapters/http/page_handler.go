package http

import (
	"context"
	"net/http"

	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/ports"
	"github.com/elchemista/driplnk/views/auth"
	"github.com/elchemista/driplnk/views/dashboard"
	"github.com/elchemista/driplnk/views/profile"
)

type PageHandler struct {
	users    domain.UserRepository
	sessions ports.SessionManager
}

func NewPageHandler(users domain.UserRepository, sessions ports.SessionManager) *PageHandler {
	return &PageHandler{
		users:    users,
		sessions: sessions,
	}
}

// Login renders the combined login/sign-up page (GitHub/Google OAuth).
func (h *PageHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if user, _ := h.currentUser(r); user != nil {
		TurboAwareRedirect(w, r, "/dashboard")
		return
	}

	if err := RenderComponent(r.Context(), w, r, auth.Login(), auth.Login()); err != nil {
		http.Error(w, "failed to render login", http.StatusInternalServerError)
	}
}

// Dashboard renders the authenticated dashboard experience (tabbed).
func (h *PageHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.currentUser(r)
	if err != nil || user == nil {
		TurboAwareRedirect(w, r, "/login")
		return
	}

	tab := r.URL.Query().Get("tab")
	if tab == "" {
		tab = "profile"
	}

	ctx := context.WithValue(r.Context(), domain.CtxKeyUser, user)
	*r = *r.WithContext(ctx)

	if err := RenderComponent(ctx, w, r, dashboard.Page(user, tab), dashboard.Frame(user, tab)); err != nil {
		http.Error(w, "failed to render dashboard", http.StatusInternalServerError)
	}
}

// Profile renders a public profile page for a given handle.
func (h *PageHandler) Profile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	handle := r.PathValue("handle")
	if handle == "" {
		http.NotFound(w, r)
		return
	}

	user, err := h.users.GetByHandle(r.Context(), handle)
	if err != nil || user == nil {
		http.NotFound(w, r)
		return
	}

	ctx := context.WithValue(r.Context(), domain.CtxKeyTargetUserID, string(user.ID))
	*r = *r.WithContext(ctx)

	if err := RenderComponent(ctx, w, r, profile.Page(user), profile.Frame(user)); err != nil {
		http.Error(w, "failed to render profile", http.StatusInternalServerError)
	}
}

func (h *PageHandler) currentUser(r *http.Request) (*domain.User, error) {
	sessionUserID, err := h.sessions.GetSession(r)
	if err != nil || sessionUserID == "" {
		return nil, err
	}

	user, err := h.users.GetByID(r.Context(), domain.UserID(sessionUserID))
	if err != nil {
		return nil, err
	}

	return user, nil
}
