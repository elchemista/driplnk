package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	adapters_http "github.com/elchemista/driplnk/internal/adapters/http"
	"github.com/elchemista/driplnk/internal/adapters/oauth"
	"github.com/elchemista/driplnk/internal/adapters/repository"
	"github.com/elchemista/driplnk/internal/adapters/seo"
	"github.com/elchemista/driplnk/internal/adapters/social"
	"github.com/elchemista/driplnk/internal/adapters/storage"
	"github.com/elchemista/driplnk/internal/config"
	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/ports"
	"github.com/elchemista/driplnk/internal/service"
	"github.com/elchemista/driplnk/views/home"
)

func main() {
	// 1. Load Server Config
	serverCfg := config.LoadServerConfig()
	log.Printf("[INFO] Starting Driplnk Server on port %s (Env: %s)", serverCfg.Port, serverCfg.Env)

	ctx := context.Background()

	// 2. Setup Storage (S3)
	s3Cfg := storage.LoadS3Config()
	var s3Store *storage.S3Store
	if s3Cfg.Bucket != "" {
		log.Printf("[INFO] Initializing S3 Store (Bucket: %s, Region: %s)", s3Cfg.Bucket, s3Cfg.Region)
		var err error
		s3Store, err = storage.NewS3Store(ctx, s3Cfg.Bucket, s3Cfg.Region)
		if err != nil {
			log.Fatalf("[FATAL] Failed to init S3 store: %v", err)
		}
	} else {
		log.Println("[INFO] S3 Bucket not configured, skipping S3 adapter")
	}

	// 3. Setup Repository (Postgres or Pebble)
	var userRepo domain.UserRepository
	var linkRepo domain.LinkRepository
	var analyticsRepo domain.AnalyticsRepository
	var dbCloser io.Closer

	// Determine which DB to use based on env (Postgres takes precedence)
	pgCfg := repository.LoadPostgresConfig()

	if pgCfg.URL != "" {
		repo, err := repository.NewPostgresRepository(pgCfg)
		if err != nil {
			log.Fatalf("[FATAL] Failed to initialize Postgres: %v", err)
		}
		userRepo = repo
		linkRepo = repository.NewPostgresLinkRepository(repo)
		analyticsRepo = repo
		dbCloser = repo
		log.Println("[INFO] Using PostgreSQL as database backend")
	} else {
		pebbleCfg := repository.LoadPebbleConfig()
		// If S3 is enabled and we are using Pebble, try restore first
		if s3Store != nil {
			log.Println("[INFO] Attempting to restore DB from S3...")
			if err := s3Store.Restore(ctx, pebbleCfg.Path); err != nil {
				log.Printf("[WARN] Failed to restore DB (fresh start or error): %v", err)
			} else {
				log.Println("[INFO] DB restored from S3 successfully")
			}
		}

		repo, err := repository.NewPebbleRepository(pebbleCfg)
		if err != nil {
			log.Fatalf("[FATAL] Failed to initialize PebbleDB: %v", err)
		}
		userRepo = repo
		linkRepo = repository.NewPebbleLinkRepository(repo)
		analyticsRepo = repo
		dbCloser = repo
		log.Println("[INFO] Using PebbleDB as database backend")
	}

	// 4. Setup Services
	// Auth Service needs Allowed Emails
	oauthCfg := oauth.LoadOAuthConfig()
	allowedEmails := config.ParseList(oauthCfg.AllowedEmails)
	log.Printf("[INFO] Initializing AuthService with %d allowed email rules", len(allowedEmails))

	authService := service.NewAuthService(userRepo, allowedEmails)

	log.Println("[INFO] Initializing AnalyticsService")
	analyticsService := service.NewAnalyticsService(analyticsRepo)

	log.Println("[INFO] Initializing HTMLFetcher for metadata extraction")
	metadataFetcher := seo.NewHTMLFetcher()

	log.Println("[INFO] Initializing LinkService with metadata fetching")
	linkService := service.NewLinkService(linkRepo, metadataFetcher)

	// 5. Setup Social Adapter (Load JSON Config)
	configDir := "config"
	var socialConfigs []config.SocialPlatformConfig
	if err := config.LoadJSONConfig(configDir+"/socials.json", &socialConfigs); err != nil {
		log.Printf("[WARN] Failed to load socials.json: %v", err)
	} else {
		log.Printf("[INFO] Loaded %d social platform configs", len(socialConfigs))
	}
	_ = social.NewSocialAdapter(socialConfigs)

	// 6. Setup OAuth Providers
	baseURL := "http://localhost:" + serverCfg.Port

	var githubProvider ports.OAuthProvider = nil
	if oauthCfg.GithubClientID != "" {
		githubProvider = oauth.NewGitHubProvider(oauthCfg, baseURL+"/auth/github/callback")
		log.Println("[INFO] GitHub OAuth Provider initialized")
	}

	var googleProvider ports.OAuthProvider = nil
	if oauthCfg.GoogleClientID != "" {
		googleProvider = oauth.NewGoogleProvider(oauthCfg, baseURL+"/auth/google/callback")
		log.Println("[INFO] Google OAuth Provider initialized")
	}

	// 7. Setup Handlers
	secureCookie := serverCfg.Port == "443"
	sessionManager := adapters_http.NewCookieSessionManager(secureCookie, "")

	authHandler := adapters_http.NewAuthHandler(authService, githubProvider, googleProvider, sessionManager, secureCookie)
	analyticsHandler := adapters_http.NewAnalyticsHandler(analyticsService)
	analyticsMiddleware := adapters_http.NewAnalyticsMiddleware(analyticsService)
	pageHandler := adapters_http.NewPageHandler(userRepo, sessionManager, linkService, analyticsService)
	userHandler := adapters_http.NewUserHandler(userRepo, sessionManager)
	linkHandler := adapters_http.NewLinkHandler(linkService, analyticsService, sessionManager, userRepo)

	// 8. HTTP Server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Auth Routes
	mux.HandleFunc("/auth/github/login", authHandler.HandleGithubLogin)
	mux.HandleFunc("/auth/github/callback", authHandler.HandleGithubCallback)
	mux.HandleFunc("/auth/google/login", authHandler.HandleGoogleLogin)
	mux.HandleFunc("/auth/google/callback", authHandler.HandleGoogleCallback)
	mux.HandleFunc("/auth/logout", authHandler.Logout)
	mux.HandleFunc("/auth/me", authHandler.Me)

	// Page Routes
	mux.HandleFunc("/login", analyticsMiddleware.TrackView(pageHandler.Login))
	mux.HandleFunc("/dashboard", analyticsMiddleware.TrackView(pageHandler.Dashboard))

	// Dashboard Profile/SEO/Theme Routes
	mux.HandleFunc("POST /dashboard/profile", userHandler.UpdateProfile)
	mux.HandleFunc("POST /dashboard/seo", userHandler.UpdateSEO)
	mux.HandleFunc("POST /dashboard/theme", userHandler.UpdateTheme)

	// Dashboard Link Routes
	mux.HandleFunc("POST /dashboard/links", linkHandler.CreateLink)
	mux.HandleFunc("POST /dashboard/links/{id}", linkHandler.UpdateLink)
	mux.HandleFunc("POST /dashboard/links/{id}/delete", linkHandler.DeleteLink)
	mux.HandleFunc("POST /dashboard/links/{id}/refresh", linkHandler.RefreshLinkMetadata)
	mux.HandleFunc("POST /dashboard/links/reorder", linkHandler.ReorderLinks)

	// Link Redirect Handler (for tracking clicks)
	mux.HandleFunc("/go/{id}", linkHandler.HandleRedirect)

	// Sitemap Handler
	sitemapHandler := adapters_http.NewSitemapHandler("http://localhost:" + serverCfg.Port)
	mux.Handle("/sitemap.xml", sitemapHandler)

	// Analytics Routes
	mux.HandleFunc("/api/analytics/scroll", analyticsHandler.RecordScroll)

	// Static Assets
	fs := http.FileServer(http.Dir("./assets/dist"))
	mux.Handle("/assets/dist/", http.StripPrefix("/assets/dist/", fs))

	// robots.txt
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./assets/robots.txt")
	})

	// Home Page (with analytics tracking)
	mux.HandleFunc("/", analyticsMiddleware.TrackView(func(w http.ResponseWriter, r *http.Request) {
		// If root path, show home page
		if r.URL.Path == "/" {
			home.Index().Render(r.Context(), w)
			return
		}

		// Otherwise, try to match as a user profile handle
		handle := r.URL.Path[1:] // Remove leading slash
		if handle != "" {
			// Try to render profile
			pageHandler.ProfileByPath(w, r, handle)
			return
		}

		adapters_http.RespondNotFound(w, r, "Page")
	}))

	// Wrap mux with middleware chain
	var handler http.Handler = mux
	handler = adapters_http.RecoveryMiddleware(handler)
	handler = adapters_http.RequestIDMiddleware(handler)

	server := &http.Server{
		Addr:    ":" + serverCfg.Port,
		Handler: handler,
	}

	// 9. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Server failed: %v", err)
		}
	}()

	<-stop
	log.Println("[INFO] Shutting down server...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Printf("[ERROR] Server shutdown error: %v", err)
	}

	if dbCloser != nil {
		dbCloser.Close()
		log.Println("[INFO] DB closed")
	}

	// Backup if using S3 and Pebble
	if s3Store != nil && pgCfg.URL == "" {
		log.Println("[INFO] Backing up DB to S3...")
		pebbleCfg := repository.LoadPebbleConfig()
		if err := s3Store.Backup(context.Background(), pebbleCfg.Path); err != nil {
			log.Printf("[ERROR] Backup failed: %v", err)
		} else {
			log.Println("[INFO] Backup successful")
		}
	}

	log.Println("[INFO] Server exited")
}
