package config

import (
	"os"
	"strings"
)

type ServerConfig struct {
	Port string
	Env  string
}

func LoadServerConfig() *ServerConfig {
	return &ServerConfig{
		Port: getEnv("PORT", "8080"),
		Env:  getEnv("ENV", "development"),
	}
}

// Helper functions (kept generic)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func ParseList(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
