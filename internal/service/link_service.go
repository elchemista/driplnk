package service

import (
	"context"
	"fmt"
	"time"

	"github.com/elchemista/driplnk/internal/domain"
	"github.com/google/uuid"
)

// LinkService encapsulates link business logic.
type LinkService struct {
	repo            domain.LinkRepository
	metadataFetcher domain.MetadataFetcher
	socialResolver  domain.SocialResolver
}

func NewLinkService(repo domain.LinkRepository, metadataFetcher domain.MetadataFetcher, socialResolver domain.SocialResolver) *LinkService {
	return &LinkService{
		repo:            repo,
		metadataFetcher: metadataFetcher,
		socialResolver:  socialResolver,
	}
}

// CreateLink creates a new link for a user.
// It automatically sets the order to be last in the list and fetches metadata.
// For social links, it resolves the platform info instead of fetching OG metadata.
func (s *LinkService) CreateLink(ctx context.Context, userID domain.UserID, title, url string, linkType domain.LinkType) (*domain.Link, error) {
	if title == "" {
		return nil, fmt.Errorf("link title is required")
	}
	if url == "" {
		return nil, fmt.Errorf("link URL is required")
	}

	// Get existing links to determine order and check for duplicates
	existingLinks, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing links: %w", err)
	}

	// Check for duplicate URL
	for _, existing := range existingLinks {
		if existing.URL == url {
			return nil, fmt.Errorf("link with this URL already exists")
		}
	}

	order := len(existingLinks) // New link goes at the end

	if linkType == "" {
		linkType = domain.LinkTypeStandard
	}

	now := time.Now()
	link := &domain.Link{
		ID:        domain.LinkID(uuid.New().String()),
		UserID:    userID,
		Title:     title,
		URL:       url,
		Type:      linkType,
		Order:     order,
		IsActive:  true,
		Metadata:  make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
	}

	// For social links, use the social resolver instead of fetching metadata
	if linkType == domain.LinkTypeSocial && s.socialResolver != nil {
		platform, err := s.socialResolver.Resolve(url)
		if err == nil && platform != nil {
			link.Metadata["social:name"] = platform.Name
			link.Metadata["social:icon"] = platform.IconSVG
			link.Metadata["social:color"] = platform.Color
			fmt.Printf("[INFO] Resolved social platform for %s: %s\n", url, platform.Name)
		} else {
			fmt.Printf("[WARN] Failed to resolve social platform for %s: %v\n", url, err)
		}
	} else if s.metadataFetcher != nil {
		// Fetch metadata from the URL for non-social links
		fmt.Printf("[INFO] Fetching metadata for URL: %s\n", url)
		metadata, err := s.metadataFetcher.Fetch(ctx, url)
		if err != nil {
			// Log the error but continue - metadata is optional
			fmt.Printf("[WARN] Failed to fetch metadata for %s: %v\n", url, err)
		} else if metadata != nil {
			// Store metadata in the link
			if metadata.Title != "" {
				link.Metadata["og:title"] = metadata.Title
			}
			if metadata.Description != "" {
				link.Metadata["og:description"] = metadata.Description
			}
			if metadata.ImageURL != "" {
				link.Metadata["og:image"] = metadata.ImageURL
			}
			fmt.Printf("[INFO] Stored metadata for %s: title=%s, has_image=%v\n",
				url, metadata.Title, metadata.ImageURL != "")
		}
	}

	if err := s.repo.Save(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to save link: %w", err)
	}

	return link, nil
}

// RefreshMetadata refetches metadata for an existing link
// For social links, it re-resolves the platform info instead of fetching OG metadata
func (s *LinkService) RefreshMetadata(ctx context.Context, linkID domain.LinkID, userID domain.UserID) (*domain.Link, error) {
	link, err := s.repo.GetByID(ctx, linkID)
	if err != nil {
		return nil, fmt.Errorf("link not found: %w", err)
	}

	// Verify ownership
	if link.UserID != userID {
		return nil, fmt.Errorf("unauthorized: link does not belong to user")
	}

	// Update metadata
	if link.Metadata == nil {
		link.Metadata = make(map[string]string)
	}

	// For social links, use the social resolver
	if link.Type == domain.LinkTypeSocial && s.socialResolver != nil {
		platform, err := s.socialResolver.Resolve(link.URL)
		if err == nil && platform != nil {
			link.Metadata["social:name"] = platform.Name
			link.Metadata["social:icon"] = platform.IconSVG
			link.Metadata["social:color"] = platform.Color
			fmt.Printf("[INFO] Refreshed social platform for %s: %s\n", link.URL, platform.Name)
		} else {
			fmt.Printf("[WARN] Failed to resolve social platform for %s: %v\n", link.URL, err)
			return nil, fmt.Errorf("failed to resolve social platform: %w", err)
		}
	} else if s.metadataFetcher != nil {
		// Fetch fresh metadata for non-social links
		fmt.Printf("[INFO] Refreshing metadata for URL: %s\n", link.URL)
		metadata, err := s.metadataFetcher.Fetch(ctx, link.URL)
		if err != nil {
			fmt.Printf("[WARN] Failed to refresh metadata for %s: %v\n", link.URL, err)
			return nil, fmt.Errorf("failed to fetch metadata: %w", err)
		}

		if metadata.Title != "" {
			link.Metadata["og:title"] = metadata.Title
		}
		if metadata.Description != "" {
			link.Metadata["og:description"] = metadata.Description
		}
		if metadata.ImageURL != "" {
			link.Metadata["og:image"] = metadata.ImageURL
		}

		fmt.Printf("[INFO] Refreshed metadata for %s: title=%s, desc=%s, image=%s\n",
			link.URL, metadata.Title, metadata.Description, metadata.ImageURL)
	}

	link.UpdatedAt = time.Now()
	if err := s.repo.Save(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to save updated metadata: %w", err)
	}

	return link, nil
}

// UpdateLink updates an existing link's fields.
func (s *LinkService) UpdateLink(ctx context.Context, linkID domain.LinkID, userID domain.UserID, title, url *string, linkType *domain.LinkType, isActive *bool) (*domain.Link, error) {
	link, err := s.repo.GetByID(ctx, linkID)
	if err != nil {
		return nil, fmt.Errorf("link not found: %w", err)
	}

	// Verify ownership
	if link.UserID != userID {
		return nil, fmt.Errorf("unauthorized: link does not belong to user")
	}

	if title != nil {
		link.Title = *title
	}
	if url != nil {
		link.URL = *url
	}
	if linkType != nil {
		link.Type = *linkType
	}
	if isActive != nil {
		link.IsActive = *isActive
	}
	link.UpdatedAt = time.Now()

	if err := s.repo.Save(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to update link: %w", err)
	}

	return link, nil
}

// DeleteLink removes a link.
func (s *LinkService) DeleteLink(ctx context.Context, linkID domain.LinkID, userID domain.UserID) error {
	link, err := s.repo.GetByID(ctx, linkID)
	if err != nil {
		return fmt.Errorf("link not found: %w", err)
	}

	if link.UserID != userID {
		return fmt.Errorf("unauthorized: link does not belong to user")
	}

	return s.repo.Delete(ctx, linkID)
}

// ReorderLinks reorders links for a user.
func (s *LinkService) ReorderLinks(ctx context.Context, userID domain.UserID, orderedIDs []domain.LinkID) error {
	return s.repo.Reorder(ctx, userID, orderedIDs)
}

// ListLinks returns all links for a user, ordered by position.
func (s *LinkService) ListLinks(ctx context.Context, userID domain.UserID) ([]*domain.Link, error) {
	return s.repo.ListByUser(ctx, userID)
}

// GetLink retrieves a single link by ID.
func (s *LinkService) GetLink(ctx context.Context, linkID domain.LinkID) (*domain.Link, error) {
	return s.repo.GetByID(ctx, linkID)
}
