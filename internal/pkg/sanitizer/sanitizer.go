package sanitizer

import (
	"strings"
)

// Normalize trims leading and trailing whitespace.
func Normalize(s string) string {
	return strings.TrimSpace(s)
}

// SanitizeInput uses Normalize for now, but can be extended for more aggressive cleaning.
func SanitizeInput(s string) string {
	return Normalize(s)
}

// SanitizeURL normalizes a URL string.
// It trims whitespace and ensures the URL starts with http:// or https://.
// If no scheme is present, it defaults to https://.
func SanitizeURL(s string) string {
	s = Normalize(s)
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return "https://" + s
	}
	return s
}
