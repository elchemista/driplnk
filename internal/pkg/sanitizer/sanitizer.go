package sanitizer

import (
	"strings"
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

var (
	strictPolicy     *bluemonday.Policy
	strictPolicyOnce sync.Once
)

func getStrictPolicy() *bluemonday.Policy {
	strictPolicyOnce.Do(func() {
		strictPolicy = bluemonday.StrictPolicy()
	})
	return strictPolicy
}

// Normalize trims leading and trailing whitespace.
func Normalize(s string) string {
	return strings.TrimSpace(s)
}

// SanitizeInput trims whitespace and removes all HTML tags to prevent XSS.
func SanitizeInput(s string) string {
	s = Normalize(s)
	return getStrictPolicy().Sanitize(s)
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
