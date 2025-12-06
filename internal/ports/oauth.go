package ports

import (
	"context"
)

type OAuthToken struct {
	AccessToken  string
	RefreshToken string
	Expiry       string
}

type OAuthUser struct {
	Email      string
	Name       string
	AvatarURL  string
	Provider   string
	ProviderID string
}

type OAuthProvider interface {
	GetAuthURL(state string) string
	Exchange(ctx context.Context, code string) (*OAuthToken, error)
	GetUserInfo(ctx context.Context, token *OAuthToken) (*OAuthUser, error)
}
