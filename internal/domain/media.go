package domain

import (
	"context"
	"io"
)

// MediaUploader defines the contract for uploading media files.
type MediaUploader interface {
	// Upload uploads the file with the given filename and returns the relative path or key.
	Upload(ctx context.Context, file io.Reader, filename string) (string, error)
	// GetURL returns the full accessible URL for the given key, handling CDN replacement if configured.
	GetURL(key string) string
}
