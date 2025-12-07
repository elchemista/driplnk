package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/elchemista/driplnk/internal/ports"
)

type GoogleProvider struct {
	config *oauth2.Config
}

func NewGoogleProvider(cfg *OAuthConfig, callbackURL string) *GoogleProvider {
	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  callbackURL,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (p *GoogleProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *GoogleProvider) Exchange(ctx context.Context, code string) (*ports.OAuthToken, error) {
	// Use the oauth2 library to exchange code
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return &ports.OAuthToken{AccessToken: token.AccessToken}, nil
}

func (p *GoogleProvider) GetUserInfo(ctx context.Context, token *ports.OAuthToken) (*ports.OAuthUser, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"https://www.googleapis.com/oauth2/v2/userinfo",
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	email, _ := result["email"].(string)
	name, _ := result["name"].(string)
	avatarURL, _ := result["picture"].(string)
	id, _ := result["id"].(string)

	return &ports.OAuthUser{
		Email:      email,
		Name:       name,
		AvatarURL:  avatarURL,
		Provider:   "google",
		ProviderID: id,
	}, nil
}
