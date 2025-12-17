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
// It currently uses Normalize, but could be extended to validate schemes, etc.
func SanitizeURL(s string) string {
	return Normalize(s)
}
