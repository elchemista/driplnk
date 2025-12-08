package main

import (
	"context"
	"fmt"
	"log"

	"github.com/elchemista/driplnk/internal/adapters/seo"
)

func main() {
	fmt.Println("Testing HTMLFetcher...")

	fetcher := seo.NewHTMLFetcher()
	ctx := context.Background()

	url := "https://elchemista.com/"
	fmt.Printf("Fetching metadata for: %s\n", url)

	metadata, err := fetcher.Fetch(ctx, url)
	if err != nil {
		log.Fatalf("Error fetching metadata: %v", err)
	}

	fmt.Printf("\n✅ SUCCESS!\n")
	fmt.Printf("Title: %s\n", metadata.Title)
	fmt.Printf("Description: %s\n", metadata.Description)
	fmt.Printf("Image URL: %s\n", metadata.ImageURL)
	fmt.Printf("URL: %s\n", metadata.URL)
}
