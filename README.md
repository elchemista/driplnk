# Driplnk

A self-hosted link management platform built with Go, designed for speed, flexibility, and privacy.

> **Status:** This project is still in progress, but most parts are done. Help or contributions are welcome.

## Screenshots

![Driplnk screenshot 1](screenshots/Schermata%20del%202025-12-23%2002-00-11.png)
![Driplnk screenshot 3](screenshots/Schermata%20del%202025-12-23%2001-59-26.png)

## Getting Started

### Prerequisites
*   Go 1.22+
*   Node.js & npm (for asset building)
*   GNU Make

### Commands

*   `make setup`: Install dependencies (Go & npm).
*   `make dev`: Start development server with hot reload (Air) and asset watching.
*   `make build`: Build production binary and assets.
*   `make generate`: Regenerate Templ files.

### Database & migrations (dev)

Dev defaults assume local Postgres with `postgres:postgres` on `localhost:5432`.

**Create the DB manually**
```bash
createdb -h localhost -U postgres driplink || psql -h localhost -U postgres -c "CREATE DATABASE driplink"
```

2) Apply the schema to `driplink`:
```bash
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/driplink?sslmode=disable
go run ./cmd/migrate up
```

3) Run the app:
```bash
source .env
make setup
make dev
```

> Production: create the database manually and run migrations against it; skip the dev-only 000000 migration.

## Configuration

Driplnk is designed to be modular. Configuration is handled via environment variables and JSON files.

### Environment Variables (.env)
| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `ENV` | Environment (`dev`, `prod`) | `dev` |
| `DATABASE_URL` | Postgres Connection String | `""` (If empty, uses Pebble) |
| `PEBBLE_PATH` | Path to Pebble DB folder | `./data/pebble` |
| `S3_BUCKET` | AWS S3 Bucket Name | `""` |
| `S3_REGION` | AWS Region | `""` |
| `S3_ACCESS_KEY` | AWS Access Key | `""` |
| `S3_SECRET_KEY` | AWS Secret Key | `""` |
| `CDN_URL` | CDN Base URL for media | `""` |
| `GITHUB_CLIENT_ID` | GitHub OAuth ID | `""` |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth Secret | `""` |
| `GOOGLE_CLIENT_ID` | Google OAuth ID | `""` |
| `GOOGLE_CLIENT_SECRET` | Google OAuth Secret | `""` |
| `ALLOWED_EMAILS` | Comma-separated list of allowed emails | `*` |

### JSON Configuration
Located in `./config/` by default.

*   **`socials.json`**: Defines supported social platforms.
    *   `id`: Internal identifier (e.g., "github").
    *   `name`: Display name.
    *   `domain`: Domain to match (e.g., "github.com").
    *   `regex`: Regex for validation/parsing.
    *   `icon`: SVG path or icon name.
    *   `color`: Brand color.

*   **`themes.json`**: Defines available themes.
    *   `id`: Theme identifier.
    *   `name`: Theme name.
    *   `background`: CSS background property.
    *   `font_family`: Font stack.
    *   `layout`: Layout type (e.g., "stack", "grid").

### Run tests

```bash
go test -v ./...
```

## Adapters & Architecture

Driplnk follows a **Hexagonal Architecture (Ports & Adapters)**.

### Core (Domain)
*   **User**: Handles user identity, profile data, and settings.
*   **Link**: Represents individual links with metadata.

### Adapters

#### 1. Repository (Database)
*   **PostgreSQL**: Primary production adapter. Uses `pgx` and standard SQL. Supports comprehensive querying.
*   **PebbleDB**: Embedded Key-Value store. Fast, single-binary deployment. Good for small instances.
    *   *Note*: When using Pebble, S3 backups are triggered on shutdown.

#### 2. Storage (Files)
*   **S3**: Stores user media (avatars, uploads) and database backups (for Pebble).
*   **Local**: (Fallback/Dev) Stores files locally.

#### 3. Authentication (OAuth)
*   **GitHub** & **Google**: Implemented via `golang.org/x/oauth2`.
*   **Session**: Cookie-based session management (`CookieSessionManager`). Secure and HttpOnly.

#### 4. Social
*   **SocialAdapter**: Resolves URLs to known social platforms using `socials.json` rules.

#### 5. HTTP
*   **Handlers**: RESTful/HTMX-ready handlers for Auth, Links, and Media.
*   **Assets**: Serves static files from `/assets/`.
*   **SEO**: Generates `robots.txt` and `sitemap.xml` dynamically.
