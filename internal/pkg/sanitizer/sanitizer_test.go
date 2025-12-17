package sanitizer

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"world", "world"},
		{"", ""},
		{"  ", ""},
		{"\t tabbed \n", "tabbed"},
	}

	for _, tt := range tests {
		if got := Normalize(tt.input); got != tt.expected {
			t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal input", "  test  ", "test"},
		{"Empty input", "", ""},
		{"Only whitespace", "   ", ""},
		{"Mixed whitespace", "\t  word  \n", "word"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeInput(tt.input); got != tt.expected {
				t.Errorf("SanitizeInput(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"With https", "https://example.com", "https://example.com"},
		{"With http", "http://example.com", "http://example.com"},
		{"Without scheme", "example.com", "https://example.com"},
		{"With whitespace", "  example.com  ", "https://example.com"},
		{"Empty string", "", ""},
		{"Only whitespace", "   ", ""},
		{"Mixed whitespace without scheme", "\t example.com \n", "https://example.com"},
		{"Subdomain without scheme", "sub.example.com/path", "https://sub.example.com/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeURL(tt.input); got != tt.expected {
				t.Errorf("SanitizeURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
