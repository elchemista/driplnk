# HOWTO Extend http adapter

Role: the HTTP adapter is the inbound edge. It owns routing-friendly handlers, middleware, and Turbo-aware helpers while delegating business logic to services and ports.

Current building blocks
- Handlers: `AuthHandler` (OAuth login/callback/logout), `PageHandler` (login/dashboard/profile templ pages), `AnalyticsHandler` (scroll API), `LinkHandler` (redirect tracking), `SitemapHandler` (XML generation).
- Middleware: `AnalyticsMiddleware.TrackView` for async view tracking.
- Helpers: `IsTurboRequest`, `TurboAwareRedirect`, `RenderComponent` for Hotwire compatibility.
- Session management: `CookieSessionManager` implements `ports.SessionManager`.

Contracts to respect
- Depend on ports/services, not concrete adapters: `service.AuthService`, `service.AnalyticsService`, `domain.UserRepository`, `domain.LinkRepository`, `ports.OAuthProvider`, `ports.SessionManager`.
- Handlers should be `http.Handler`/`http.HandlerFunc` compatible and context-friendly (see `r.PathValue`, `context.WithValue` usage).
- Avoid embedding domain logic; validate/shape inputs, then call services.

How to add/extend handlers
1) Define a new struct with injected dependencies via a constructor (e.g. `NewFooHandler(repo domain.UserRepository, svc *service.BarService)`).
2) Expose `ServeHTTP` or focused methods (e.g. `HandleCreate`, `HandleUpdate`) that operate on `http.ResponseWriter`/`*http.Request`.
3) Reuse helpers: set Turbo-friendly redirects, use `RenderComponent` for templ partials, and wrap routes with `AnalyticsMiddleware` if the page should emit view events.
4) Keep sessions behind `ports.SessionManager`; avoid setting cookies manually outside that port.
5) Register the route in `cmd/server/main.go` (or another entrypoint) by wiring the handler into the `http.ServeMux`.

End-to-end workflow
Request → optional middleware (`TrackView`) → handler (parses input, pulls session/user, renders templ or redirects) → service call → repository/storage/OAuth adapter → persistence or external API → response (Turbo-aware if needed).

Testing tips
- Use `httptest` with fake ports/services; existing tests (`analytics_handler_test.go`, `sitemap_test.go`) show patterns for validating status codes, headers, and payloads.
