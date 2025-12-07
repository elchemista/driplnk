# HOWTO Extend seo adapter

Role: supply metadata fetching behind the `domain.MetadataFetcher` port used to enrich links with OpenGraph/Twitter data.

Port contract (`internal/domain/metadata.go`)
- `Fetch(ctx context.Context, url string) (*domain.LinkMetadata, error)` returning `Title`, `Description`, `ImageURL`, and `URL`.

Current adapter
- `HTMLFetcher`: HTTP client with a 10s timeout that requests the page, parses HTML, and extracts OG/Twitter meta tags plus a fallback `<title>`.

How to build another fetcher
1) Decide on the data source (cache layer, headless browser, 3rd-party API). Create a struct that keeps the necessary client(s) and implements `Fetch`.  
2) Preserve the output shape: always set `URL` and fill `Title`/`Description`/`ImageURL` when available; return typed errors on network/parse failures.  
3) Be polite to targets: set a user agent, add timeouts, and consider rate limits.  
4) If adding caching, keep it outside the domain: wrap a real fetcher and short-circuit when cached.  
5) Test with fixtures for HTML variants and error cases.

Workflow integration
- Inject your `MetadataFetcher` into the service or handler that builds links/products. The flow becomes: handler receives a URL → service calls fetcher to enrich metadata → repository saves the link with `Metadata` populated → HTTP layer renders richer cards.
