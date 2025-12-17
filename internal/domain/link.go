package domain

import (
	"context"
	"time"
)

type LinkID string
type LinkType string

const (
	LinkTypeStandard LinkType = "standard"
	LinkTypeSocial   LinkType = "social"
	LinkTypeProduct  LinkType = "product"
)

type Link struct {
	ID         LinkID            `json:"id"`
	UserID     UserID            `json:"user_id"`
	Title      string            `json:"title" validate:"required,max=100"`
	URL        string            `json:"url" validate:"required,url"`
	Type       LinkType          `json:"type" validate:"oneof=standard social product"`
	Order      int               `json:"order"`
	IsActive   bool              `json:"is_active"`
	Metadata   map[string]string `json:"metadata,omitempty"` // Stores icon_name, og:title, og:image, etc.
	ClickCount uint64            `json:"click_count"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type LinkRepository interface {
	Save(ctx context.Context, link *Link) error
	GetByID(ctx context.Context, id LinkID) (*Link, error)
	ListByUser(ctx context.Context, userID UserID) ([]*Link, error)
	Delete(ctx context.Context, id LinkID) error
	Reorder(ctx context.Context, userID UserID, linkIDs []LinkID) error
}
