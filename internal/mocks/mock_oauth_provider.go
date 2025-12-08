package mocks

import (
	"context"
	"fmt"

	"github.com/elchemista/driplnk/internal/ports"
)

// MockOAuthProvider is a test double for ports.OAuthProvider.
type MockOAuthProvider struct {
	// Configurable responses
	AuthURL     string
	Token       *ports.OAuthToken
	User        *ports.OAuthUser
	ExchangeErr error
	UserInfoErr error

	// Track method calls
	GetAuthURLCalls  []string
	ExchangeCalls    []string
	GetUserInfoCalls int
}

func NewMockOAuthProvider() *MockOAuthProvider {
	return &MockOAuthProvider{
		AuthURL: "https://mock.oauth.example/auth",
		Token: &ports.OAuthToken{
			AccessToken:  "mock-access-token",
			RefreshToken: "mock-refresh-token",
		},
		User: &ports.OAuthUser{
			Email:      "test@example.com",
			Name:       "Test User",
			AvatarURL:  "https://example.com/avatar.png",
			Provider:   "mock",
			ProviderID: "mock-12345",
		},
		GetAuthURLCalls: make([]string, 0),
		ExchangeCalls:   make([]string, 0),
	}
}

func (m *MockOAuthProvider) GetAuthURL(state string) string {
	m.GetAuthURLCalls = append(m.GetAuthURLCalls, state)
	if m.AuthURL == "" {
		return fmt.Sprintf("https://mock.oauth.example/auth?state=%s", state)
	}
	return fmt.Sprintf("%s?state=%s", m.AuthURL, state)
}

func (m *MockOAuthProvider) Exchange(ctx context.Context, code string) (*ports.OAuthToken, error) {
	m.ExchangeCalls = append(m.ExchangeCalls, code)
	if m.ExchangeErr != nil {
		return nil, m.ExchangeErr
	}
	return m.Token, nil
}

func (m *MockOAuthProvider) GetUserInfo(ctx context.Context, token *ports.OAuthToken) (*ports.OAuthUser, error) {
	m.GetUserInfoCalls++
	if m.UserInfoErr != nil {
		return nil, m.UserInfoErr
	}
	return m.User, nil
}

// SetUser is a helper to configure the user returned from OAuth.
func (m *MockOAuthProvider) SetUser(email, name, avatarURL, provider, providerID string) {
	m.User = &ports.OAuthUser{
		Email:      email,
		Name:       name,
		AvatarURL:  avatarURL,
		Provider:   provider,
		ProviderID: providerID,
	}
}
