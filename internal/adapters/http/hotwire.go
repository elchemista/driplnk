package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/a-h/templ"
)

// IsTurboRequest detects Turbo-driven navigations or frame visits so handlers can
// respond with Turbo-friendly redirects or partials.
func IsTurboRequest(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "turbo-stream") ||
		r.Header.Get("Turbo-Frame") != "" ||
		r.Header.Get("Turbo-Visit") != ""
}

// IsTurboFrameRequest is true when the request originated from a <turbo-frame>.
func IsTurboFrameRequest(r *http.Request) bool {
	if r.Header.Get("Turbo-Frame") != "" {
		return true
	}
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "turbo-stream")
}

// TurboAwareRedirect performs a redirect that plays nicely with Turbo Drive/Frames.
// Turbo interprets the Turbo-Location header even on non-3xx responses.
func TurboAwareRedirect(w http.ResponseWriter, r *http.Request, location string) {
	if IsTurboRequest(r) {
		w.Header().Set("Turbo-Location", location)
		// 303 ensures Turbo performs a GET follow-up even when the request was not idempotent.
		w.WriteHeader(http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, location, http.StatusSeeOther)
}

// RenderComponent renders a templ component, automatically falling back to the frame
// version when the request originated from a Turbo Frame.
func RenderComponent(ctx context.Context, w http.ResponseWriter, r *http.Request, full templ.Component, frame templ.Component) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if IsTurboFrameRequest(r) && frame != nil {
		return frame.Render(ctx, w)
	}

	return full.Render(ctx, w)
}
