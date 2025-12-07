# HOWTO Extend storage adapter

Role: handle media uploads and database backups behind simple ports so the app stays cloud-agnostic.

Current adapters
- `S3Uploader`: implements `domain.MediaUploader` with `Upload(ctx, file, filename) (key, error)` and `GetURL(key) string`. Supports optional `FolderPath` prefix and CDN URL rewriting.
- `S3Store`: helper for Pebble backups/restores. `Backup(ctx, localPath)` zips a directory and uploads `driplnk_backup.zip`; `Restore(ctx, localPath)` downloads and unzips it.
- Config: `LoadS3Config` reads `S3_BUCKET`, `S3_REGION`, `CDN_URL`.

How to add another storage backend
1) Implement `domain.MediaUploader`: keep the same semantics (return a stable key, provide a URL builder). Create a config struct/env loader for credentials/paths.  
2) If you need backups for a different DB, mirror the `S3Store` API or expose a small interface for backup/restore and implement it for your provider.  
3) Use `context.Context` for all network calls and set sensible timeouts. Handle retryable errors explicitly.  
4) Keep uploads pure (no global env reads inside methods); inject everything through constructors like `NewXYZUploader(cfg)` for testability.  
5) Add tests with local fakes or wire in a mock client to avoid real cloud calls.

Workflow integration
- `cmd/server/main.go` currently instantiates `S3Store` (for Pebble backup/restore at startup/shutdown). Wire your uploader/store the same way and inject it into the services/handlers that accept `domain.MediaUploader` when media endpoints are added.
