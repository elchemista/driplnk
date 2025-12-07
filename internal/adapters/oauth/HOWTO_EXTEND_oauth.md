# HOWTO Extend oauth adapter

Role: provide OAuth providers that satisfy `ports.OAuthProvider` for the auth flow used by `AuthHandler`.

Required interface (`internal/ports/oauth.go`)
- `GetAuthURL(state string) string`: build provider login URL with CSRF `state`.
- `Exchange(ctx, code string) (*ports.OAuthToken, error)`: swap code for access token.
- `GetUserInfo(ctx, token *ports.OAuthToken) (*ports.OAuthUser, error)`: fetch email/name/avatar/provider ID.

Current providers
- `GitHubProvider` and `GoogleProvider` wrap `golang.org/x/oauth2` with provider-specific endpoints and scopes.
- Config comes from `OAuthConfig` (`GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `ALLOWED_EMAILS`).

How to add a new provider
1) Extend `OAuthConfig` with client ID/secret env vars for the new provider and load them in `LoadOAuthConfig`.  
2) Create a `XYZProvider` struct holding an `oauth2.Config` (or custom client) and ensure it implements the three methods above. Include any scopes and callback URL.  
3) Normalize the returned user fields to `ports.OAuthUser` (email, display name, avatar URL, provider string, provider ID).  
4) Wire it in `cmd/server/main.go`: instantiate the provider with the correct callback URL, pass it to `NewAuthHandler`, and add login/callback routes.  
5) Add tests with a fake HTTP server to assert token exchange and user info parsing.

Workflow integration
`AuthHandler` → provider `GetAuthURL` (login redirect) → provider `Exchange` + `GetUserInfo` on callback → `AuthService.LoginOrRegister` → `SessionManager.CreateSession` → Turbo redirect to dashboard.
