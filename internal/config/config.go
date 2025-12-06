package config

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

type Config struct {
	Port               string
	DBPath             string
	S3Bucket           string
	S3Region           string
	CDNURL             string
	AllowedEmails      []string // Parsed from comma-separated env
	GoogleClientID     string
	GoogleClientSecret string
	GithubClientID     string
	GithubClientSecret string
	Socials            []SocialPlatformConfig
	Themes             ThemeConfig
}

type SocialPlatformConfig struct {
	Name         string `json:"name"`
	Domain       string `json:"domain"`
	RegexPattern string `json:"regex_pattern"`
	IconSVG      string `json:"icon_svg"`
	Color        string `json:"color"`
}

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

func Load() *Config {
	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		DBPath:             getEnv("DB_PATH", "./data"),
		S3Bucket:           getEnv("S3_BUCKET", ""),
		S3Region:           getEnv("S3_REGION", "us-east-1"),
		CDNURL:             getEnv("CDN_URL", ""),
		AllowedEmails:      parseList(getEnv("ALLOWED_EMAILS", "*")),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GithubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GithubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
	}

	// Allow overriding config directory via env (useful for tests/deployments)
	configDir := getEnv("CONFIG_DIR", "config")
	loadJSONConfig(configDir+"/socials.json", &cfg.Socials)
	loadJSONConfig(configDir+"/themes.json", &cfg.Themes)

	return cfg
}

func loadJSONConfig(path string, target interface{}) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Warning: failed to read config file %s: %v", path, err)
		return
	}
	if err := json.Unmarshal(data, target); err != nil {
		log.Printf("Warning: failed to parse config file %s: %v", path, err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func parseList(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func (c *Config) IsEmailAllowed(email string) bool {
	for _, allowed := range c.AllowedEmails {
		if allowed == "*" || allowed == email {
			return true
		}
	}
	return false
}
