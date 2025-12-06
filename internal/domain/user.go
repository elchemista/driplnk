package domain

import (
	"context"
	"time"
)

type UserID string

type SEOMeta struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
}

type User struct {
	ID          UserID    `json:"id"`
	Email       string    `json:"email"`
	Handle      string    `json:"handle"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	SEOMeta     SEOMeta   `json:"seo_meta,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Repository interfaces define the contract for data persistence
// adhering to Hexagonal Architecture (Port).
type UserRepository interface {
	Save(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id UserID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByHandle(ctx context.Context, handle string) (*User, error)
}
