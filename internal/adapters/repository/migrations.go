package repository

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// ApplyMigrations uses golang-migrate to apply migrations from the migrations directory
// Note: We need to adapt the embedded FS path since we moved files to root `migrations/`
// and this file is in `internal/adapters/repository`.
//
// However, since we moved the physical files to `{ROOT}/migrations`, we can't embed them
// easily from here without go modules support or using a wrapper.
//
// STRATEGY CHANGE:
// Use the CLI for primary migrations.
// But to keep existing behavior (migration on startup), user can run the CLI before server.
// OR we can make the server run it.
//
// Given we moved files to root, `//go:embed` here won't work on `../../migrations` easily in standard toolchain unless module root.
//
// Let's assume for now we want the Server to run migrations using the library if possible.
// But since the CLI is the requested way, maybe we disable automatic migration in the app code
// OR strictly require the `migrations` folder to be present if running from binary without embed.
//
// Safest bet for "clean code" requested by user:
// 1. Rely on CLI for migrations.
// 2. Remove automatic migration from `postgres.go` or make it optional/check version.
//
// Reverting to: Automatic migration is NICE, but if files are external, we need to read them from FS.
//
// Let's implement ApplyMigrations that reads from `file://migrations` relative to CWD.
func ApplyMigrations(dbURL string) error {
	m, err := migrate.New(
		"file://migrations",
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Migrations already up to date")
			return nil
		}
		return err
	}
	log.Println("Migrations applied successfully")
	return nil
}
