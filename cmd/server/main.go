package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elchemista/driplnk/internal/adapters/repository"
	"github.com/elchemista/driplnk/internal/adapters/storage"
	"github.com/elchemista/driplnk/internal/config"
	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/service"
)

func main() {
	// 1. Load Config
	cfg := config.Load()
	log.Println("Config loaded")

	ctx := context.Background()

	// 2. Setup S3 Store (if configured)
	var s3Store *storage.S3Store
	if cfg.S3Bucket != "" {
		var err error
		s3Store, err = storage.NewS3Store(ctx, cfg.S3Bucket, cfg.S3Region)
		if err != nil {
			log.Fatalf("Failed to init S3 store: %v", err)
		}
		log.Println("S3 Store initialized")

		// 3. Attempt Restore from S3 (before opening DB)
		log.Println("Attempting to restore DB from S3...")
		if err := s3Store.Restore(ctx, cfg.DBPath); err != nil {
			log.Printf("Warning: Failed to restore DB from S3 (might be fresh start): %v", err)
		} else {
			log.Println("DB restored from S3 successfully")
		}
	} else {
		log.Println("S3 Bucket not configured, skipping S3 backup/restore")
	}

	// 4. Initialize Pebble DB
	repo, err := repository.NewPebbleRepository(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open Pebble DB: %v", err)
	}
	defer repo.Close()
	log.Println("Pebble DB opened")

	// 5. Initialize Services
	// Cast repo to interface to ensure compliance
	var userRepo domain.UserRepository = repo
	// var linkRepo domain.LinkRepository = repo // Unused yet

	authService := service.NewAuthService(userRepo, cfg)

	// 6. Setup HTTP Server
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// TODO: Add Auth Handlers and Link Handlers
	_ = authService

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	// 7. Graceful Shutdown & Backup
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Close DB explicitly before backup
	if err := repo.Close(); err != nil {
		log.Printf("Error closing DB: %v", err)
	} else {
		log.Println("Pebble DB closed")
	}

	// 8. S3 Backup
	if s3Store != nil {
		log.Println("Backing up DB to S3...")
		if err := s3Store.Backup(context.Background(), cfg.DBPath); err != nil {
			log.Printf("Error backing up to S3: %v", err)
		} else {
			log.Println("DB backed up to S3 successfully")
		}
	}

	log.Println("Server exited")
}
