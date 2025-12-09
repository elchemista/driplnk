# Driplnk

## Features

* Login with Google & GitHub
* Personalized links & unlimited links
* Personalized style and layout
* Analytics
* Products & affiliate links
* Socials (with Icons) - **Configurable via JSON**
* Subscriptions
* Theme Customization (Background, Animations, Layout, Fonts, Colors) - **Configurable via JSON**

## Tech Stack

* Go
* Hotwire
* Tailwind (DaisyUI)
* **Dual Database Support**:
    * Embedded DB: Pebble
    * Relational DB: PostgreSQL (optimized for Neon)
* Migration System: `golang-migrate`

## Hosting

* Fly.io / Neon (Postgres) / AWS S3 (Backups & Media)

---

## Code Style

* Follow SOLID principles.
* Use Dependency Injection where it makes sense.
* Follow Clean Code project structure.
* Write self-contained tests.
* Test coverage > 80%.

---

## Idea

This project is a link manager similar to Linktree but designed with self-hosting in mind.
It uses Golang + Hotwire + Tailwind + (Pebble DB or Postgres) to create a flexible link management system.

## Architecture

The project follows **Hexagonal Architecture (Ports and Adapters)** to decouple core logic from infrastructure.

### Structure
*   **`cmd`**:
    *   **`server`**: Entry point. Wires Config, Storage (Pebble/Postgres/S3), and Services using Dependency Injection.
    *   **`migrate`**: CLI tool for managing PostgreSQL migrations.
*   **`internal/domain`**: Pure business entities (`User`, `Link`) and Repository Interfaces (`UserRepository`). No external dependencies.
*   **`internal/service`**: Business logic implementations (`AuthService`). Orchestrates data flow.
*   **`internal/ports`**: Domain interfaces/ports (e.g., `OAuthProvider`).
*   **`internal/config`**: Modular configuration structs (`ServerConfig`, `PostgresConfig`, etc.) and JSON loaders.
*   **`internal/adapters`**:
    *   **`repository`**: `PebbleRepository` & `PostgresRepository`.
    *   **`oauth`**: GitHub and Google implementations of `OAuthProvider`.
    *   **`http`**: HTTP Handlers (`AuthHandler`, etc.).
    *   **`storage`**: `S3Store` for backups and media.
*   **`views`**: Templ templates for the UI.

### Data Flow
1.  **Startup**: `main.go` loads modular Configs -> initializes Adapters (Postgres/Pebble, S3, OAuth) -> injects into Services -> injects into Handlers.
2.  **Request**: Handler -> Service (`AuthService`) -> Domain Logic.
3.  **Persistence**: Service -> Repository Interface -> Adapter Implementation -> DB.

## Implementation Status

### Completed
*   [x] **Project Structure**: Set up following Clean/Hexagonal patterns.
*   [x] **Domain Modeling**: `User` and `Link` entities defined.
*   [x] **Storage Engines**:
    *   `PebbleDB` integration for high-performance KV storage.
    *   `PostgreSQL` integration with connection pooling (Neon-ready).
    *   **Migration CLI**: `cmd/migrate` for managing SQL schemas.
*   [x] **Repository Layer**:
    *   Generic JSON serialization for entities (Pebble).
    *   SQL implementations with optimized indexes (Postgres).
*   [x] **Backup System**: Automated S3 backup on shutdown (Pebble only).
*   [x] **Auth Service**: Basic logic for user registration/login and handle uniqueness.
*   [x] **Server Skeleton**: HTTP server wiring with graceful shutdown.
*   [x] **User Extensions**:
    *   Added `Title`, `Description`, and `SEOMeta` to `User` domain.
    *   [x] Added `Theme` struct (Background, Animations, Layout, Colors, Fonts).
*   [x] **OAuth System**:
    *   `OAuthProvider` port.
    *   Adapters for GitHub and Google using `golang.org/x/oauth2`.
    *   HTTP Handlers with CSRF protection (state cookie) and session management.
*   [x] **Media System**:
    *   `MediaUploader` interface.
    *   S3 Adapter with CDN support (`CDN_URL`).
*   [x] **Social Integration**:
    *   `SocialResolver` interface.
    *   `SocialAdapter` utilizing regex for platform detection.
*   [x] **SEO & Metadata**:
    *   `MetadataFetcher` interface.
    *   `HTMLFetcher` adapter for OpenGraph/Twitter card parsing.
*   [x] **Configuration System**:
    *   **Modular Architecture**: Decoupled configuration structs (`PostgresConfig`, `ServerConfig`, etc.).
    *   **Dependency Injection**: Components accept only the config they need.
    *   **JSON Loaders**: For `Socials` and `Themes`.
*   [x] **Session Management**:
    *   `SessionManager` interface (Ports).
    *   `CookieSessionManager` implementation (Adapters).
    *   Secure/HttpOnly cookie configuration.
*   [x] **Serving Essentials**:
    *   `SitemapHandler` for dynamic XML generation.
    *   `robots.txt` serving from assets.
    *   Static asset serving (`/assets/`).
*   [x] **Frontend Infrastructure**:
    *   `esbuild` for JavaScript bundling.
    *   `tailwindcss` (v4) via `@tailwindcss/cli`.
    *   `Hotwire` (Turbo & Stimulus) integration.
    *   `Templ` for type-safe Go templates.
*   [x] **Hotwire Pages**:
    *   Turbo-aware auth/logout redirects and helpers.
    *   Login page with GitHub/Google OAuth.
    *   Dashboard with Turbo Frame tabs (profile, links, theme, analytics).
    *   Public profile page templated with theme preview.

### Pending
*   [ ] Test Suite: `LinkHandler`, `UserHandler`, `PageHandler`.

## Current Work Plan

1.  **Frontend**: Implement visitor tracking (Completed).
2.  **Tests**: Implement comprehensive test suite for handlers.
3.  **Documentation**: Finalize project documentation.
