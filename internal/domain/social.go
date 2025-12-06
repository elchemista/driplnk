package domain

// SocialPlatform represents a detected social media platform and its styling properties.
type SocialPlatform struct {
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	IconSVG string `json:"icon_svg"`
	Color   string `json:"color"`
	Effect  string `json:"effect,omitempty"`
}

// SocialResolver defines the contract for resolving a URL to a widely known social platform.
type SocialResolver interface {
	// Resolve returns a SocialPlatform if the url matches a known platform, or nil if not found.
	Resolve(url string) (*SocialPlatform, error)
}
