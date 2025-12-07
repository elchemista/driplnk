package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSitemapHandler_ServeHTTP(t *testing.T) {
	handler := NewSitemapHandler("http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/xml" {
		t.Errorf("expected Content-Type application/xml, got %s", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "<urlset") {
		t.Error("expected body to contain <urlset")
	}
	if !strings.Contains(body, "http://localhost:8080/") {
		t.Error("expected body to contain base URL")
	}
	if !strings.Contains(body, "http://localhost:8080/login") {
		t.Error("expected body to contain /login")
	}
}
