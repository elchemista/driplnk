package config

// SocialPlatformConfig defines configuration for social media platforms
type SocialPlatformConfig struct {
	Name         string `json:"name"`
	Domain       string `json:"domain"`
	RegexPattern string `json:"regex_pattern"`
	IconSVG      string `json:"icon_svg"`
	Color        string `json:"color"`
}

// ThemeConfig defines available theme options
type ThemeConfig struct {
	Layouts     []ThemeOption `json:"layouts"`
	Backgrounds []ThemeOption `json:"backgrounds"`
	Patterns    []ThemeOption `json:"patterns"`
	Fonts       []ThemeOption `json:"fonts"`
	Animations  []ThemeOption `json:"animations"`
}

type ThemeOption struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type,omitempty"`
	Value    string `json:"value,omitempty"`
	CSSClass string `json:"css_class,omitempty"`
	Family   string `json:"family,omitempty"`
}
