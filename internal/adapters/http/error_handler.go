package http

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/elchemista/driplnk/internal/domain"
	viewerrors "github.com/elchemista/driplnk/views/errors"
)

// ErrorHandler provides methods for handling and rendering errors.
type ErrorHandler struct{}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// HandleError handles an error and renders the appropriate error page.
func (h *ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	// Extract AppError if present
	appErr := domain.GetAppError(err)
	if appErr == nil {
		// Wrap unknown errors as internal errors
		appErr = domain.NewInternalError(err.Error())
	}

	// Log the error with context
	h.logError(r, appErr)

	// Set HTTP status code
	w.WriteHeader(appErr.Code)

	// Render appropriate error template
	h.renderErrorPage(w, r, appErr.Code, appErr.Message)
}

// logError logs the error with request context.
func (h *ErrorHandler) logError(r *http.Request, appErr *domain.AppError) {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	logMsg := fmt.Sprintf("[ERROR] RequestID=%s Method=%s Path=%s Code=%d Message=%s",
		requestID,
		r.Method,
		r.URL.Path,
		appErr.Code,
		appErr.Message,
	)

	if appErr.Details != "" {
		logMsg += fmt.Sprintf(" Details=%s", appErr.Details)
	}

	// For 5xx errors, include stack trace
	if appErr.Code >= 500 {
		log.Printf("%s\nStack:\n%s", logMsg, string(debug.Stack()))
	} else {
		log.Println(logMsg)
	}
}

// renderErrorPage renders the appropriate error template.
func (h *ErrorHandler) renderErrorPage(w http.ResponseWriter, r *http.Request, code int, message string) {
	ctx := r.Context()

	switch code {
	case 400:
		viewerrors.Error400().Render(ctx, w)
	case 401:
		viewerrors.Error401().Render(ctx, w)
	case 403:
		viewerrors.Error403().Render(ctx, w)
	case 404:
		viewerrors.Error404().Render(ctx, w)
	case 409:
		viewerrors.Error409().Render(ctx, w)
	case 429:
		viewerrors.Error429().Render(ctx, w)
	case 500:
		viewerrors.Error500().Render(ctx, w)
	case 502:
		viewerrors.Error502().Render(ctx, w)
	case 503:
		viewerrors.Error503().Render(ctx, w)
	case 504:
		viewerrors.Error504().Render(ctx, w)
	default:
		// For other codes, use custom error
		title := http.StatusText(code)
		if title == "" {
			title = "Error"
		}
		viewerrors.CustomError(code, title, message, "").Render(ctx, w)
	}
}

// Helper functions for common error responses

// RespondNotFound renders a 404 error page.
func RespondNotFound(w http.ResponseWriter, r *http.Request, resource string) {
	LogError(r, 404, fmt.Sprintf("%s not found", resource), "")
	w.WriteHeader(http.StatusNotFound)
	viewerrors.Error404().Render(r.Context(), w)
}

// RespondUnauthorized renders a 401 error page.
func RespondUnauthorized(w http.ResponseWriter, r *http.Request, message string) {
	if message == "" {
		message = "Authentication required"
	}
	LogError(r, 401, message, "")
	w.WriteHeader(http.StatusUnauthorized)
	viewerrors.Error401().Render(r.Context(), w)
}

// RespondForbidden renders a 403 error page.
func RespondForbidden(w http.ResponseWriter, r *http.Request, message string) {
	if message == "" {
		message = "Permission denied"
	}
	LogError(r, 403, message, "")
	w.WriteHeader(http.StatusForbidden)
	viewerrors.Error403().Render(r.Context(), w)
}

// RespondBadRequest renders a 400 error page or returns JSON/Turbo error.
func RespondBadRequest(w http.ResponseWriter, r *http.Request, message string) {
	LogError(r, 400, message, "")

	if IsTurboRequest(r) {
		w.Header().Set("Content-Type", "text/vnd.turbo-stream.html")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `<turbo-stream action="update" target="flash-messages">
			<template>
				<div class="alert alert-error">
					<span>%s</span>
				</div>
			</template>
		</turbo-stream>`, message)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	viewerrors.Error400().Render(r.Context(), w)
}

// RespondConflict renders a 409 error page.
func RespondConflict(w http.ResponseWriter, r *http.Request, message string) {
	LogError(r, 409, message, "")

	if IsTurboRequest(r) {
		w.Header().Set("Content-Type", "text/vnd.turbo-stream.html")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, `<turbo-stream action="update" target="flash-messages">
			<template>
				<div class="alert alert-error">
					<span>%s</span>
				</div>
			</template>
		</turbo-stream>`, message)
		return
	}

	w.WriteHeader(http.StatusConflict)
	viewerrors.Error409().Render(r.Context(), w)
}

// RespondInternalError renders a 500 error page.
func RespondInternalError(w http.ResponseWriter, r *http.Request, err error) {
	details := ""
	if err != nil {
		details = err.Error()
	}
	LogError(r, 500, "Internal server error", details)
	log.Printf("[ERROR] Stack trace:\n%s", string(debug.Stack()))

	w.WriteHeader(http.StatusInternalServerError)
	viewerrors.Error500().Render(r.Context(), w)
}

// RespondTooManyRequests renders a 429 error page.
func RespondTooManyRequests(w http.ResponseWriter, r *http.Request) {
	LogError(r, 429, "Rate limit exceeded", "")
	w.WriteHeader(http.StatusTooManyRequests)
	viewerrors.Error429().Render(r.Context(), w)
}

// LogError logs an error with request context.
func LogError(r *http.Request, code int, message, details string) {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	userID := ""
	if uid := r.Context().Value(domain.CtxKeyUser); uid != nil {
		if user, ok := uid.(*domain.User); ok {
			userID = string(user.ID)
		}
	}

	logMsg := fmt.Sprintf("[ERROR] RequestID=%s Code=%d Method=%s Path=%s",
		requestID, code, r.Method, r.URL.Path)

	if userID != "" {
		logMsg += fmt.Sprintf(" UserID=%s", userID)
	}

	logMsg += fmt.Sprintf(" Message=%q", message)

	if details != "" {
		logMsg += fmt.Sprintf(" Details=%q", details)
	}

	log.Println(logMsg)
}

// LogInfo logs an informational message with request context.
func LogInfo(r *http.Request, message string) {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	log.Printf("[INFO] RequestID=%s Method=%s Path=%s Message=%s",
		requestID, r.Method, r.URL.Path, message)
}

// LogWarn logs a warning message with request context.
func LogWarn(r *http.Request, message string) {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	log.Printf("[WARN] RequestID=%s Method=%s Path=%s Message=%s",
		requestID, r.Method, r.URL.Path, message)
}

// RecoveryMiddleware catches panics and renders 500 error pages.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				err := fmt.Errorf("panic: %v", rec)
				LogError(r, 500, "Panic recovered", err.Error())
				log.Printf("[PANIC] Stack trace:\n%s", string(debug.Stack()))

				w.WriteHeader(http.StatusInternalServerError)
				viewerrors.Error500().Render(r.Context(), w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware adds a unique request ID to each request.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
			r.Header.Set("X-Request-ID", requestID)
		}
		w.Header().Set("X-Request-ID", requestID)

		ctx := context.WithValue(r.Context(), "requestID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// NotFoundHandler returns a custom 404 handler.
func NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		LogError(r, 404, "Page not found", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		viewerrors.Error404().Render(r.Context(), w)
	}
}

// MethodNotAllowedHandler returns a 405 handler.
func MethodNotAllowedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		LogError(r, 405, "Method not allowed", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		viewerrors.CustomError(405, "Method Not Allowed",
			fmt.Sprintf("The %s method is not allowed for this resource.", r.Method),
			"").Render(r.Context(), w)
	}
}

// Helper to check for specific errors

// IsNotFoundError checks if error is a not found error.
func IsNotFoundError(err error) bool {
	return errors.Is(err, domain.ErrNotFound)
}

// IsUnauthorizedError checks if error is an unauthorized error.
func IsUnauthorizedError(err error) bool {
	return errors.Is(err, domain.ErrUnauthorized)
}

// IsForbiddenError checks if error is a forbidden error.
func IsForbiddenError(err error) bool {
	return errors.Is(err, domain.ErrForbidden)
}
