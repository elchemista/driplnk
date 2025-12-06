package service

import (
	"context"
	"errors"
	"time"

	"github.com/elchemista/driplnk/internal/config"
	"github.com/elchemista/driplnk/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrUserNotAllowed = errors.New("user email is not allowed")
)

type AuthService struct {
	userRepo domain.UserRepository
	config   *config.Config
}

func NewAuthService(userRepo domain.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		config:   cfg,
	}
}

// LoginOrRegister handles the OAuth callback logic
// It checks if email is allowed, creates user if new, or returns existing.
func (s *AuthService) LoginOrRegister(ctx context.Context, email, handle, avatarURL string) (*domain.User, error) {
	if !s.config.IsEmailAllowed(email) {
		return nil, ErrUserNotAllowed
	}

	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return existingUser, nil
	}

	// User not found, create new
	// If handle is empty or taken, we might need logic to generate one.
	// For now, assuming handle comes from OAuth (like GitHub username).
	// Ideally we should verify uniqueness.
	
	// Check handle uniqueness
	if _, err := s.userRepo.GetByHandle(ctx, handle); err == nil {
		// Handle taken. For simplicity, append random string or UUID
		handle = handle + "-" + uuid.New().String()[:4]
	}

	newUser := &domain.User{
		ID:        domain.UserID(uuid.New().String()),
		Email:     email,
		Handle:    handle,
		AvatarURL: avatarURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Save(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}
