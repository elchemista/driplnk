package http

import (
	"encoding/xml"
	"net/http"
	"time"
)

type URL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

type UrlSet struct {
	XMLName string `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	URLs    []URL  `xml:"url"`
}

type SitemapHandler struct {
	BaseURL string
	// In the future, we can inject a service to fetch dynamic routes (e.g. user profiles)
}

func NewSitemapHandler(baseURL string) *SitemapHandler {
	return &SitemapHandler{
		BaseURL: baseURL,
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
		// Add more static pages here
	}

	// 2. Dynamic Routes (TODO: Fetch from DB)
	// for _, user := range users {
	// 	urls = append(urls, URL{Loc: h.BaseURL + "/" + user.Handle ...})
	// }

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
