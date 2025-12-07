# HOWTO Extend social adapter

Role: resolve arbitrary URLs to known social platforms behind the `domain.SocialResolver` port so links can display the right branding/icon.

Port contract (`internal/domain/social.go`)
- `Resolve(url string) (*domain.SocialPlatform, error)`: return platform info (name, domain, icon SVG, color) or an error when not matched.

Current adapter
- `SocialAdapter`: compiles regex rules from `config.SocialPlatformConfig` and returns the first matching platform.
- Config is typically loaded from `config/socials.json` in `cmd/server/main.go`.

How to extend
1) Add or tweak platform configs (name/domain/regex/icon/color) in the JSON file and ensure the regex is valid. Invalid patterns are skipped during construction.  
2) If you need smarter matching (e.g., URL parsing, hostname lists), adjust `platformRule` creation or add another matcher function before the regex loop.  
3) Keep returns immutable: copy the platform before returning so callers cannot mutate shared state.  
4) Consider adding tests with representative URLs to assert resolution order and failure cases.

Workflow integration
- Config loader → `NewSocialAdapter` → stored resolver injected wherever links are built/enriched → UI uses the resolved `SocialPlatform` to show the correct icon/color next to the link.
