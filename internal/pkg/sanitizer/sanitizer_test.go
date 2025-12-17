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
	// Currently same as Normalize
	if got := SanitizeInput("  test  "); got != "test" {
		t.Errorf("SanitizeInput failed: got %q", got)
	}
}
