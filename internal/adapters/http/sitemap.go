package http

import (
	"context"
	"encoding/xml"
	"log"
	"net/http"
	"time"

	"github.com/elchemista/driplnk/internal/domain"
)

type URL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

type UrlSet struct {
	XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	URLs    []URL    `xml:"url"`
}

type SitemapHandler struct {
	BaseURL  string
	UserRepo domain.UserRepository
}

func NewSitemapHandler(baseURL string, userRepo domain.UserRepository) *SitemapHandler {
	return &SitemapHandler{
		BaseURL:  baseURL,
		UserRepo: userRepo,
	}
}

func (h *SitemapHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Static Routes
	urls := []URL{
		{
			Loc:        h.BaseURL + "/",
			LastMod:    time.Now().Format("2006-01-02"),
			ChangeFreq: "daily",
			Priority:   "1.0",
		},
		{
			Loc:        h.BaseURL + "/login",
			ChangeFreq: "monthly",
			Priority:   "0.8",
		},
	}

	// 2. Dynamic Routes - Fetch all users from repository
	if h.UserRepo != nil {
		users, err := h.UserRepo.ListAll(context.Background())
		if err != nil {
			log.Printf("[WARN] Failed to list users for sitemap: %v", err)
		} else {
			for _, user := range users {
				if user.Handle != "" {
					urls = append(urls, URL{
						Loc:        h.BaseURL + "/" + user.Handle,
						LastMod:    user.UpdatedAt.Format("2006-01-02"),
						ChangeFreq: "weekly",
						Priority:   "0.6",
					})
				}
			}
		}
	}

	// 3. Render XML
	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(xml.Header))

	urlSet := UrlSet{URLs: urls}
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(urlSet); err != nil {
		http.Error(w, "Failed to generate sitemap", http.StatusInternalServerError)
	}
}
