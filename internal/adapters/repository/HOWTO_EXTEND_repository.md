# HOWTO Extend repository adapter

Role: persist domain data behind the repository ports. Each adapter must satisfy the domain interfaces so services stay storage-agnostic.

Ports to implement (`internal/domain`)
- `UserRepository`: `Save`, `GetByID`, `GetByEmail`, `GetByHandle`.
- `LinkRepository`: `Save`, `GetByID`, `ListByUser`, `Delete`, `Reorder`.
- `AnalyticsRepository`: `SaveEvent`, `GetSummary`.
- Reuse `ErrNotFound` semantics for missing rows/keys.

Current adapters
- `PostgresRepository`: SQL-backed, applies migrations via `ApplyMigrations`, implements all three ports and `Close()`. Connection tuned via `PostgresConfig`.
- `PebbleRepository`: embedded KV store implementing all three ports (with a no-op `Reorder`) and `Close()`. Uses JSON serialization.
- `ApplyMigrations`: runs `golang-migrate` against `file://migrations`.

How to add a new persistence backend
1) Choose which ports you need to support; implement all three if you want feature parity with Postgres.  
2) Create a constructor (`NewXYZRepository(cfg *XYZConfig)`) that opens the connection, applies any schema/bootstrap, and returns something that satisfies the ports (and `io.Closer` if applicable).  
3) Implement the methods with `context.Context` awareness. Normalize “not found” to `ErrNotFound` so handlers/services can branch consistently.  
4) Keep serialization stable (`json` for Pebble) and ensure link/user metadata is marshaled/unmarshaled exactly like existing adapters.  
5) Add config structs/loaders similar to `LoadPostgresConfig`/`LoadPebbleConfig` and expose env-driven options.  
6) Write tests that exercise the interface, not just the concrete type (fake DB, dockerized DB, or in-memory KV).

Workflow integration
- `cmd/server/main.go` currently picks Postgres when `DATABASE_URL` is set, otherwise Pebble; both get injected into `AuthService` and `AnalyticsService` (and future Link services). Swap in your adapter by constructing it there and assigning it to the same port variables.
- Call `Close()` during shutdown and (if needed) pair with backup/restore adapters (see `storage` S3Store) before exit.
