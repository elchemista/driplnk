package http

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/elchemista/driplnk/internal/ports"
	"github.com/elchemista/driplnk/internal/service"
)

type AuthHandler struct {
	authService    *service.AuthService
	github         ports.OAuthProvider
	google         ports.OAuthProvider
	sessionManager ports.SessionManager
	secure         bool
}

func NewAuthHandler(authService *service.AuthService, github, google ports.OAuthProvider, sessionManager ports.SessionManager, secure bool) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		github:         github,
		google:         google,
		sessionManager: sessionManager,
		secure:         secure,
	}
}

// generateState creates a random string to prevent CSRF
func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (h *AuthHandler) HandleGithubLogin(w http.ResponseWriter, r *http.Request) {
	if h.github == nil {
		http.Error(w, "GitHub OAuth is not configured", http.StatusServiceUnavailable)
		return
	}
	state := generateState()
	// In production, store state in a secure, HttpOnly cookie with expiration
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   h.secure,
		Path:     "/",
	})
	http.Redirect(w, r, h.github.GetAuthURL(state), http.StatusTemporaryRedirect)
}

func (h *AuthHandler) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.google == nil {
		http.Error(w, "Google OAuth is not configured", http.StatusServiceUnavailable)
		return
	}
	state := generateState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   h.secure,
		Path:     "/",
	})
	http.Redirect(w, r, h.google.GetAuthURL(state), http.StatusTemporaryRedirect)
}

func (h *AuthHandler) HandleGithubCallback(w http.ResponseWriter, r *http.Request) {
	if h.github == nil {
		http.Error(w, "GitHub OAuth is not configured", http.StatusServiceUnavailable)
		return
	}
	h.handleCallback(w, r, h.github)
}

func (h *AuthHandler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.google == nil {
		http.Error(w, "Google OAuth is not configured", http.StatusServiceUnavailable)
		return
	}
	h.handleCallback(w, r, h.google)
}

func (h *AuthHandler) handleCallback(w http.ResponseWriter, r *http.Request, provider ports.OAuthProvider) {
	// 1. Validate State
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		http.Error(w, "State cookie missing", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("state") != stateCookie.Value {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// 2. Exchange Code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code missing", http.StatusBadRequest)
		return
	}

	token, err := provider.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Token exchange failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 3. Get User Info
	oauthUser, err := provider.GetUserInfo(r.Context(), token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
		return
	}

	// 4. Login/Register in Domain
	user, err := h.authService.LoginOrRegister(r.Context(), oauthUser.Email, oauthUser.Name, oauthUser.AvatarURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Login failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 5. Create Session
	if err := h.sessionManager.CreateSession(r.Context(), w, string(user.ID)); err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Redirect to dashboard
	TurboAwareRedirect(w, r, "/dashboard")
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear session
	if err := h.sessionManager.ClearSession(r.Context(), w); err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	if IsTurboRequest(r) {
		TurboAwareRedirect(w, r, "/")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out"))
}

// Debug handler to check session
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := h.sessionManager.GetSession(r)
	if err != nil || userID == "" {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	// Ideally we fetch user from DB here
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"user_id": userID})
}
