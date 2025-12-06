# Driplnk

## Features

* Login with google & github
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
* Embedded DB Pebble

## Hostring

* Fly.io

---

## Code style

* follow SOLID principles
* Use only where it make sense dependencies injection
* Also follow clean code project structure
* Write self contained test 
* Test should cover at least 80% of the code

---

## Idea

This project is a link manager similar to linktree but in self hosted version in mind. 
It use golang + hotwire + tailwind + pebble db to create a self hosted link manager system.

## Architecture

The project follows **Hexagonal Architecture (Ports and Adapters)** to decouple core logic from infrastructure.

### Structure
*   **`cmd/server`**: Entry point. Wires Config, Storage (Pebble/S3), and Services.
*   **`internal/domain`**: Pure business entities (`User`, `Link`) and Repository Interfaces (`UserRepository`). No external dependencies.
*   **`internal/service`**: Business logic implementations (`AuthService`). Orchestrates data flow.
*   **`internal/ports`**: Domain interfaces/ports (e.g., `OAuthProvider`).
*   **`internal/adapters`**:
    *   **`repository`**: `PebbleRepository` (Adapter for `UserRepository`).
    *   **`oauth`**: GitHub and Google implementations of `OAuthProvider`.
    *   **`http`**: HTTP Handlers (`AuthHandler`, etc.).
    *   **`storage`**: `S3Store` for backups.
*   **`views`**: Templ templates for the UI.

### Data Flow
1.  **Startup**: `main.go` loads Config -> initializes `S3Store` (restores DB) -> opens `PebbleDB`.
2.  **Request**: Handler -> Service (`AuthService`) -> Domain Logic.
3.  **Persistence**: Service -> Repository Interface -> `PebbleRepository` (JSON serialization) -> Disk.
4.  **Backup**: On shutdown, `S3Store` snapshots the DB to S3.

## Implementation Status

### Completed
*   [x] **Project Structure**: Set up following Clean/Hexagonal patterns.
*   [x] **Domain Modeling**: `User` and `Link` entities defined.
*   [x] **Storage Engine**: `PebbleDB` integration for high-performance KV storage.
*   [x] **Repository Layer**:
    *   Generic JSON serialization for entities.
    *   Indexing strategies (e.g., `user:email:idx`).
*   [x] **Backup System**: Automated S3 backup on shutdown and restore on startup.
*   [x] **Auth Service**: Basic logic for user registration/login and handle uniqueness.
*   [x] **Server Skeleton**: HTTP server wiring with graceful shutdown.

*   [x] **User Extensions**:
    *   Added `Title`, `Description`, and `SEOMeta` to `User` domain.
    *   [x] Added `Theme` struct (Background, Animations, Layout, Colors, Fonts).
*   [x] **OAuth System**:
    *   `OAuthProvider` port.
    *   Adapters for GitHub and Google.
    *   HTTP Handlers with CSRF protection (state cookie) and session management.
*   [x] **Media System**:
    *   `MediaUploader` interface.
    *   S3 Adapter with CDN support (`CDN_URL`).
*   [x] **Social Integration**:
    *   `SocialResolver` interface.
    *   `SocialAdapter` utilizing regex for platform detection (Facebook, Twitter, etc.).
*   [x] **SEO & Metadata**:
    *   `MetadataFetcher` interface.
    *   `HTMLFetcher` adapter for OpenGraph/Twitter card parsing.
*   [x] **Configuration System**:
    *   JSON-based config for `Socials` and `Themes`.
    *   Dynamic loading and regex compilation.
    *   Environment variable support for config path.

### Pending
*   [ ] Link CRUD Handlers.
*   [ ] Frontend Templates (Login, Dashboard, Public Profile).
*   [ ] HTMX Integration.
