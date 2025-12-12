package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCSRFMiddleware_BlocksWithoutToken(t *testing.T) {
	handler := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}), false)

	// POST without CSRF token should be blocked
	req := httptest.NewRequest(http.MethodPost, "/some-endpoint", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden, got %d", rec.Code)
	}
}

func TestCSRFMiddleware_AllowsGET(t *testing.T) {
	handler := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}), false)

	// GET should pass without token
	req := httptest.NewRequest(http.MethodGet, "/some-endpoint", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for GET, got %d", rec.Code)
	}
}

func TestCSRFMiddleware_AllowsMatchingToken(t *testing.T) {
	handler := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}), false)

	// First, make a GET to obtain the token from cookie
	getReq := httptest.NewRequest(http.MethodGet, "/", nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)

	// Extract the csrf_token cookie
	cookies := getRec.Result().Cookies()
	var csrfToken string
	for _, c := range cookies {
		if c.Name == "csrf_token" {
			csrfToken = c.Value
			break
		}
	}

	if csrfToken == "" {
		t.Fatal("CSRF cookie not set on GET request")
	}

	// Now make a POST with the token in header and cookie
	postReq := httptest.NewRequest(http.MethodPost, "/some-endpoint", nil)
	postReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})
	postReq.Header.Set("X-CSRF-Token", csrfToken)
	postRec := httptest.NewRecorder()

	handler.ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for POST with valid token, got %d: %s", postRec.Code, postRec.Body.String())
	}
}

func TestCSRFMiddleware_BlocksMismatchedToken(t *testing.T) {
	handler := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}), false)

	// POST with mismatched tokens
	postReq := httptest.NewRequest(http.MethodPost, "/some-endpoint", nil)
	postReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: "correct-token"})
	postReq.Header.Set("X-CSRF-Token", "different-token")
	postRec := httptest.NewRecorder()

	handler.ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for mismatched token, got %d", postRec.Code)
	}
}
