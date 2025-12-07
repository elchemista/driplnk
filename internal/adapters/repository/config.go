package repository

import (
	"os"
	"time"
)

type PostgresConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func LoadPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		URL:             os.Getenv("DATABASE_URL"),
		MaxOpenConns:    25, // Could be env var
		MaxIdleConns:    5,  // Could be env var
		ConnMaxLifetime: 5 * time.Minute,
	}
}

type PebbleConfig struct {
	Path string
}

func LoadPebbleConfig() *PebbleConfig {
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "./data"
	}
	return &PebbleConfig{
		Path: path,
	}
}
