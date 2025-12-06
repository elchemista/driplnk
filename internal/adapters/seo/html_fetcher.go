package seo

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/elchemista/driplnk/internal/domain"
	"golang.org/x/net/html"
)

type HTMLFetcher struct {
	client *http.Client
}

func NewHTMLFetcher() *HTMLFetcher {
	return &HTMLFetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (f *HTMLFetcher) Fetch(ctx context.Context, url string) (*domain.LinkMetadata, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Driplnk-Bot/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %w", err)
	}

	meta := &domain.LinkMetadata{URL: url}
	extractMeta(doc, meta)

	return meta, nil
}

func extractMeta(n *html.Node, meta *domain.LinkMetadata) {
	if n.Type == html.ElementNode && n.Data == "meta" {
		var property, name, content string
		for _, attr := range n.Attr {
			switch attr.Key {
			case "property":
				property = attr.Val
			case "name":
				name = attr.Val
			case "content":
				content = attr.Val
			}
		}

		if property == "og:title" || name == "twitter:title" {
			if meta.Title == "" {
				meta.Title = content
			}
		}
		if property == "og:description" || name == "twitter:description" || name == "description" {
			if meta.Description == "" {
				meta.Description = content
			}
		}
		if property == "og:image" || name == "twitter:image" {
			if meta.ImageURL == "" {
				meta.ImageURL = content
			}
		}
	}

	// Fallback for title tag if og:title not found
	if n.Type == html.ElementNode && n.Data == "title" && meta.Title == "" {
		if n.FirstChild != nil {
			meta.Title = n.FirstChild.Data
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractMeta(c, meta)
	}
}
