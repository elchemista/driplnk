package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/elchemista/driplnk/internal/config"
	"github.com/elchemista/driplnk/internal/ports"
)

type GoogleProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewGoogleProvider(cfg *config.Config, redirectURL string) *GoogleProvider {
	return &GoogleProvider{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  redirectURL,
	}
}

func (p *GoogleProvider) GetAuthURL(state string) string {
	return fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email%%20profile&state=%s",
		p.ClientID,
		url.QueryEscape(p.RedirectURL),
		state,
	)
}

func (p *GoogleProvider) Exchange(ctx context.Context, code string) (*ports.OAuthToken, error) {
	requestBodyMap := url.Values{}
	requestBodyMap.Set("client_id", p.ClientID)
	requestBodyMap.Set("client_secret", p.ClientSecret)
	requestBodyMap.Set("code", code)
	requestBodyMap.Set("grant_type", "authorization_code")
	requestBodyMap.Set("redirect_uri", p.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = requestBodyMap.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if errVal, ok := result["error"]; ok {
		return nil, fmt.Errorf("google error: %v, desc: %v", errVal, result["error_description"])
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get access token from response")
	}

	return &ports.OAuthToken{
		AccessToken: accessToken,
	}, nil
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
