package mocks

import (
	"github.com/elchemista/driplnk/internal/domain"
)

// MockSocialResolver is a test double for domain.SocialResolver.
type MockSocialResolver struct {
	// Configurable responses
	Platform   *domain.SocialPlatform
	ResolveErr error
	Platforms  map[string]*domain.SocialPlatform // URL substring -> platform

	// Track method calls
	ResolveCalls []string
}

func NewMockSocialResolver() *MockSocialResolver {
	return &MockSocialResolver{
		Platforms:    make(map[string]*domain.SocialPlatform),
		ResolveCalls: make([]string, 0),
	}
}

func (m *MockSocialResolver) Resolve(url string) (*domain.SocialPlatform, error) {
	m.ResolveCalls = append(m.ResolveCalls, url)

	if m.ResolveErr != nil {
		return nil, m.ResolveErr
	}

	// Check if URL matches any configured platform
	for pattern, platform := range m.Platforms {
		if contains(url, pattern) {
			return platform, nil
		}
	}

	// Return default platform if set
	if m.Platform != nil {
		return m.Platform, nil
	}

	return nil, nil
}

// AddPlatform configures a URL pattern to resolve to a platform.
func (m *MockSocialResolver) AddPlatform(urlPattern string, platform *domain.SocialPlatform) {
	m.Platforms[urlPattern] = platform
}

// SetDefaultPlatform sets the fallback platform for unmatched URLs.
func (m *MockSocialResolver) SetDefaultPlatform(name, domainName, icon, color string) {
	m.Platform = &domain.SocialPlatform{
		Name:    name,
		Domain:  domainName,
		IconSVG: icon,
		Color:   color,
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
