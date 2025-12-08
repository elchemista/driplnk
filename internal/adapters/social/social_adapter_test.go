package social_test

import (
	"testing"

	"github.com/elchemista/driplnk/internal/adapters/social"
	"github.com/elchemista/driplnk/internal/config"
)

func TestSocialAdapter_Resolve(t *testing.T) {
	configs := []config.SocialPlatformConfig{
		{
			Name:         "GitHub",
			Domain:       "github.com",
			RegexPattern: `github\.com`,
			IconSVG:      "<svg>github</svg>",
			Color:        "#333",
		},
		{
			Name:         "Twitter",
			Domain:       "twitter.com",
			RegexPattern: `twitter\.com|x\.com`,
			IconSVG:      "<svg>twitter</svg>",
			Color:        "#1DA1F2",
		},
		{
			Name:         "YouTube",
			Domain:       "youtube.com",
			RegexPattern: `youtube\.com`,
			IconSVG:      "<svg>youtube</svg>",
			Color:        "#FF0000",
		},
	}

	adapter := social.NewSocialAdapter(configs)

	t.Run("resolves GitHub URL", func(t *testing.T) {
		platform, err := adapter.Resolve("https://github.com/elchemista")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if platform == nil {
			t.Fatal("expected platform, got nil")
		}
		if platform.Name != "GitHub" {
			t.Errorf("expected GitHub, got %s", platform.Name)
		}
		if platform.Color != "#333" {
			t.Errorf("expected #333, got %s", platform.Color)
		}
	})

	t.Run("resolves Twitter URL", func(t *testing.T) {
		platform, err := adapter.Resolve("https://twitter.com/username")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if platform == nil {
			t.Fatal("expected platform, got nil")
		}
		if platform.Name != "Twitter" {
			t.Errorf("expected Twitter, got %s", platform.Name)
		}
	})

	t.Run("resolves X.com as Twitter", func(t *testing.T) {
		platform, err := adapter.Resolve("https://x.com/username")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if platform == nil {
			t.Fatal("expected platform, got nil")
		}
		if platform.Name != "Twitter" {
			t.Errorf("expected Twitter, got %s", platform.Name)
		}
	})

	t.Run("returns error for unknown URL", func(t *testing.T) {
		platform, err := adapter.Resolve("https://unknown-site.xyz/page")
		// Adapter returns error when platform not found
		if err == nil {
			t.Error("expected error for unknown URL")
		}
		if platform != nil {
			t.Errorf("expected nil platform for unknown URL, got %s", platform.Name)
		}
	})

	t.Run("returns error for empty URL", func(t *testing.T) {
		platform, err := adapter.Resolve("")
		// Adapter returns error when no match
		if err == nil {
			t.Error("expected error for empty URL")
		}
		if platform != nil {
			t.Error("expected nil for empty URL")
		}
	})
}

func TestSocialAdapter_EmptyConfig(t *testing.T) {
	adapter := social.NewSocialAdapter(nil)

	t.Run("returns error for any URL with empty config", func(t *testing.T) {
		platform, err := adapter.Resolve("https://github.com/test")
		// With empty config, no rules match, so error is returned
		if err == nil {
			t.Error("expected error with empty config")
		}
		if platform != nil {
			t.Error("expected nil with empty config")
		}
	})
}

func TestSocialAdapter_InvalidRegex(t *testing.T) {
	configs := []config.SocialPlatformConfig{
		{
			Name:         "Invalid",
			RegexPattern: "[invalid(regex", // Invalid regex - will be skipped
		},
		{
			Name:         "Valid",
			Domain:       "valid.com",
			RegexPattern: `valid\.com`,
		},
	}

	// Should not panic, just skip invalid regex
	adapter := social.NewSocialAdapter(configs)

	t.Run("still resolves valid platforms", func(t *testing.T) {
		platform, err := adapter.Resolve("https://valid.com/page")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if platform == nil || platform.Name != "Valid" {
			t.Error("expected Valid platform to still work")
		}
	})
}
