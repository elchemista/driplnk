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
	repo domain.LinkRepository
}

func NewLinkService(repo domain.LinkRepository) *LinkService {
	return &LinkService{repo: repo}
}

// CreateLink creates a new link for a user.
// It automatically sets the order to be last in the list.
func (s *LinkService) CreateLink(ctx context.Context, userID domain.UserID, title, url string, linkType domain.LinkType) (*domain.Link, error) {
	if title == "" {
		return nil, fmt.Errorf("link title is required")
	}
	if url == "" {
		return nil, fmt.Errorf("link URL is required")
	}

	// Get existing links to determine order
	existingLinks, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing links: %w", err)
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

	if err := s.repo.Save(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to save link: %w", err)
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
