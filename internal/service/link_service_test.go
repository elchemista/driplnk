package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/mocks"
	"github.com/elchemista/driplnk/internal/service"
)

func TestLinkService_CreateLink(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockLinkRepository()
	metadataFetcher := mocks.NewMockMetadataFetcher()
	svc := service.NewLinkService(repo, metadataFetcher, nil)
	userID := domain.UserID("user-123")

	t.Run("creates link successfully", func(t *testing.T) {
		link, err := svc.CreateLink(ctx, userID, "My Portfolio", "https://example.com", domain.LinkTypeStandard)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if link.ID == "" {
			t.Error("expected link ID to be set")
		}
		if link.Title != "My Portfolio" {
			t.Errorf("expected title 'My Portfolio', got %s", link.Title)
		}
		if link.URL != "https://example.com" {
			t.Errorf("expected URL 'https://example.com', got %s", link.URL)
		}
		if link.Type != domain.LinkTypeStandard {
			t.Errorf("expected type standard, got %s", link.Type)
		}
		if !link.IsActive {
			t.Error("expected link to be active")
		}
	})

	t.Run("returns error for empty title", func(t *testing.T) {
		_, err := svc.CreateLink(ctx, userID, "", "https://example.com", domain.LinkTypeStandard)
		if err == nil {
			t.Error("expected error for empty title")
		}
	})

	t.Run("returns error for empty URL", func(t *testing.T) {
		_, err := svc.CreateLink(ctx, userID, "Title", "", domain.LinkTypeStandard)
		if err == nil {
			t.Error("expected error for empty URL")
		}
	})

	t.Run("sets order based on existing links", func(t *testing.T) {
		// Add some existing links
		repo.AddLink(&domain.Link{
			ID:     "existing-1",
			UserID: userID,
			Order:  0,
		})
		repo.AddLink(&domain.Link{
			ID:     "existing-2",
			UserID: userID,
			Order:  1,
		})

		link, err := svc.CreateLink(ctx, userID, "New Link", "https://new.example.com", domain.LinkTypeStandard)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		// Should be at the end (3 links now including previous test)
		if link.Order < 2 {
			t.Errorf("expected order >= 2, got %d", link.Order)
		}
	})
}

func TestLinkService_UpdateLink(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockLinkRepository()
	metadataFetcher := mocks.NewMockMetadataFetcher()
	svc := service.NewLinkService(repo, metadataFetcher, nil)
	userID := domain.UserID("user-123")
	linkID := domain.LinkID("link-456")

	// Setup existing link
	repo.AddLink(&domain.Link{
		ID:        linkID,
		UserID:    userID,
		Title:     "Original Title",
		URL:       "https://original.com",
		Type:      domain.LinkTypeStandard,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	t.Run("updates title", func(t *testing.T) {
		newTitle := "Updated Title"
		link, err := svc.UpdateLink(ctx, linkID, userID, &newTitle, nil, nil, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if link.Title != newTitle {
			t.Errorf("expected title '%s', got '%s'", newTitle, link.Title)
		}
	})

	t.Run("updates URL", func(t *testing.T) {
		newURL := "https://updated.com"
		link, err := svc.UpdateLink(ctx, linkID, userID, nil, &newURL, nil, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if link.URL != newURL {
			t.Errorf("expected URL '%s', got '%s'", newURL, link.URL)
		}
	})

	t.Run("updates active status", func(t *testing.T) {
		inactive := false
		link, err := svc.UpdateLink(ctx, linkID, userID, nil, nil, nil, &inactive)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if link.IsActive != inactive {
			t.Errorf("expected IsActive=%v, got %v", inactive, link.IsActive)
		}
	})

	t.Run("rejects update from different user", func(t *testing.T) {
		otherUser := domain.UserID("other-user")
		newTitle := "Hacked Title"
		_, err := svc.UpdateLink(ctx, linkID, otherUser, &newTitle, nil, nil, nil)
		if err == nil {
			t.Error("expected error for unauthorized user")
		}
	})
}

func TestLinkService_DeleteLink(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockLinkRepository()
	metadataFetcher := mocks.NewMockMetadataFetcher()
	svc := service.NewLinkService(repo, metadataFetcher, nil)
	userID := domain.UserID("user-123")
	linkID := domain.LinkID("link-789")

	repo.AddLink(&domain.Link{
		ID:     linkID,
		UserID: userID,
	})

	t.Run("deletes link successfully", func(t *testing.T) {
		err := svc.DeleteLink(ctx, linkID, userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify deleted
		_, err = repo.GetByID(ctx, linkID)
		if err == nil {
			t.Error("expected link to be deleted")
		}
	})

	t.Run("rejects delete from different user", func(t *testing.T) {
		anotherLink := domain.LinkID("another-link")
		repo.AddLink(&domain.Link{
			ID:     anotherLink,
			UserID: userID,
		})

		otherUser := domain.UserID("other-user")
		err := svc.DeleteLink(ctx, anotherLink, otherUser)
		if err == nil {
			t.Error("expected error for unauthorized user")
		}
	})
}

func TestLinkService_ListLinks(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockLinkRepository()
	metadataFetcher := mocks.NewMockMetadataFetcher()
	svc := service.NewLinkService(repo, metadataFetcher, nil)
	userID := domain.UserID("user-list")

	// Add links
	repo.AddLink(&domain.Link{ID: "link-1", UserID: userID, Order: 2})
	repo.AddLink(&domain.Link{ID: "link-2", UserID: userID, Order: 0})
	repo.AddLink(&domain.Link{ID: "link-3", UserID: userID, Order: 1})

	t.Run("returns links sorted by order", func(t *testing.T) {
		links, err := svc.ListLinks(ctx, userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(links) != 3 {
			t.Fatalf("expected 3 links, got %d", len(links))
		}
		if links[0].ID != "link-2" || links[1].ID != "link-3" || links[2].ID != "link-1" {
			t.Error("links not sorted by order")
		}
	})

	t.Run("returns empty for unknown user", func(t *testing.T) {
		links, err := svc.ListLinks(ctx, domain.UserID("unknown"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(links) != 0 {
			t.Errorf("expected 0 links, got %d", len(links))
		}
	})
}
