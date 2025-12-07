# HOWTO Extend handler adapter

Purpose: keep non-HTTP drivers (cron jobs, CLI flows, queue/webhook consumers) that trigger domain services without adding infrastructure concerns inside `internal/service` or `internal/domain`.

What a handler should look like
- Small struct that owns only the ports/services it needs (e.g. `*service.AnalyticsService`, `domain.UserRepository`).
- Explicit entrypoints such as `Handle(ctx context.Context, payload SomeEvent) error` or `Run(ctx)`; avoid global state.
- Pure orchestration: validate/shape input, call services, emit logs/metrics; no business rules duplicated here.

How to add a new handler
1) Identify the trigger (scheduled job, message topic, webhook) and design a minimal payload/port for it. If you need a new port, add it under `internal/ports` first.  
2) Create a handler struct in this folder that depends on the existing ports/services rather than concrete adapters.  
3) Implement a clear method (`Handle`, `ServeHTTP` wrapper, or `Run`) that wires the trigger to the service call chain.  
4) Register the handler in the appropriate entrypoint (`cmd/server` for long-lived processes, a new command in `cmd/*` for batch jobs) and inject the already-configured adapters (DB, storage, OAuth, etc.).  
5) Cover the orchestration with focused tests by faking the ports/services.

Workflow integration
- External event → handler in this folder → service layer (`internal/service`) → repositories/storage/OAuth adapters → persistence or outbound calls.
- Keep logging/metrics at the handler edge; push domain decisions down to services so HTTP and background paths stay consistent.
