package domain

import "context"

// LinkMetadata contains SEO and OpenGraph data extracted from a URL.
type LinkMetadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	URL         string `json:"url"`
}

// MetadataFetcher defines the contract for fetching SEO metadata from a URL.
type MetadataFetcher interface {
	// Fetch retrieves metadata for the given URL.
	Fetch(ctx context.Context, url string) (*LinkMetadata, error)
}
