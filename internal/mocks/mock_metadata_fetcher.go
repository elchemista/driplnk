package mocks

import (
	"context"

	"github.com/elchemista/driplnk/internal/domain"
)

// MockMetadataFetcher is a test double for domain.MetadataFetcher.
type MockMetadataFetcher struct {
	// Configurable responses
	Metadata      *domain.LinkMetadata
	FetchErr      error
	MetadataByURL map[string]*domain.LinkMetadata

	// Track method calls
	FetchCalls []string
}

func NewMockMetadataFetcher() *MockMetadataFetcher {
	return &MockMetadataFetcher{
		MetadataByURL: make(map[string]*domain.LinkMetadata),
		FetchCalls:    make([]string, 0),
	}
}

func (m *MockMetadataFetcher) Fetch(ctx context.Context, url string) (*domain.LinkMetadata, error) {
	m.FetchCalls = append(m.FetchCalls, url)

	if m.FetchErr != nil {
		return nil, m.FetchErr
	}

	// Check URL-specific metadata
	if metadata, ok := m.MetadataByURL[url]; ok {
		return metadata, nil
	}

	// Return default metadata
	if m.Metadata != nil {
		return m.Metadata, nil
	}

	return &domain.LinkMetadata{
		Title:       "Default Title",
		Description: "Default Description",
		URL:         url,
	}, nil
}

// AddMetadata configures metadata for a specific URL.
func (m *MockMetadataFetcher) AddMetadata(url string, metadata *domain.LinkMetadata) {
	m.MetadataByURL[url] = metadata
}

// SetDefaultMetadata sets the fallback metadata for unmatched URLs.
func (m *MockMetadataFetcher) SetDefaultMetadata(title, description, imageURL string) {
	m.Metadata = &domain.LinkMetadata{
		Title:       title,
		Description: description,
		ImageURL:    imageURL,
	}
}
