package mocks

import (
	"context"
	"io"
)

// MockMediaUploader is a test double for domain.MediaUploader.
type MockMediaUploader struct {
	// Configurable responses
	UploadKey string
	UploadErr error
	BaseURL   string

	// Track method calls
	UploadCalls []struct {
		Filename string
		Content  []byte
	}
	GetURLCalls []string
}

func NewMockMediaUploader() *MockMediaUploader {
	return &MockMediaUploader{
		UploadKey: "media/uploaded-file.png",
		BaseURL:   "https://cdn.example.com",
		UploadCalls: make([]struct {
			Filename string
			Content  []byte
		}, 0),
		GetURLCalls: make([]string, 0),
	}
}

func (m *MockMediaUploader) Upload(ctx context.Context, file io.Reader, filename string) (string, error) {
	content, _ := io.ReadAll(file)
	m.UploadCalls = append(m.UploadCalls, struct {
		Filename string
		Content  []byte
	}{filename, content})

	if m.UploadErr != nil {
		return "", m.UploadErr
	}

	if m.UploadKey != "" {
		return m.UploadKey, nil
	}
	return "media/" + filename, nil
}

func (m *MockMediaUploader) GetURL(key string) string {
	m.GetURLCalls = append(m.GetURLCalls, key)
	return m.BaseURL + "/" + key
}

// SetUploadResponse configures the key returned from upload.
func (m *MockMediaUploader) SetUploadResponse(key string) {
	m.UploadKey = key
}

// SetBaseURL configures the CDN URL for GetURL.
func (m *MockMediaUploader) SetBaseURL(baseURL string) {
	m.BaseURL = baseURL
}
