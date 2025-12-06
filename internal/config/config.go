package config

import (
	"os"
	"strings"
)

type Config struct {
	Port          string
	DBPath        string
	S3Bucket      string
	S3Region      string
	AllowedEmails []string // Parsed from comma-separated env
	GoogleClientID     string
	GoogleClientSecret string
	GithubClientID     string
	GithubClientSecret string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		DBPath:             getEnv("DB_PATH", "./data"),
		S3Bucket:           getEnv("S3_BUCKET", ""),
		S3Region:           getEnv("S3_REGION", "us-east-1"),
		AllowedEmails:      parseList(getEnv("ALLOWED_EMAILS", "*")),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GithubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GithubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
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
