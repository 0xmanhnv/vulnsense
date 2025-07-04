package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
	"vulnsense/internal/adapter"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Test with a known RSS feed
	testFeeds := []struct {
		name     string
		url      string
		feedType string
		source   string
	}{
		{
			name:     "BleepingComputer",
			url:      "https://www.bleepingcomputer.com/feed/",
			feedType: "news",
			source:   "BleepingComputer",
		},
		{
			name:     "KrebsOnSecurity",
			url:      "https://krebsonsecurity.com/feed",
			feedType: "blog",
			source:   "KrebsOnSecurity",
		},
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	for _, feed := range testFeeds {
		fmt.Printf("\n=== Testing %s ===\n", feed.name)

		// Create RSS adapter
		rssAdapter := adapter.NewRSSFeedAdapter(
			feed.name,
			feed.url,
			feed.feedType,
			feed.source,
			httpClient,
			logger,
		)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

		vulnerabilities, err := rssAdapter.Fetch(ctx)
		if err != nil {
			fmt.Printf("Error fetching from %s: %v\n", feed.name, err)
			cancel()
			continue
		}

		fmt.Printf("Found %d potential vulnerabilities from %s\n", len(vulnerabilities), feed.name)

		// Print first few vulnerabilities
		for i, vuln := range vulnerabilities {
			if i >= 3 { // Only show first 3
				break
			}
			fmt.Printf("  %d. %s [%s] - %s\n", i+1, vuln.ID, vuln.Severity, vuln.Title)
			fmt.Printf("     Product: %s | CVEs: %v\n", vuln.ProductName, vuln.CVEIDs)
			fmt.Printf("     Tags: %v\n", vuln.Tags)
		}

		if len(vulnerabilities) > 3 {
			fmt.Printf("  ... and %d more\n", len(vulnerabilities)-3)
		}

		cancel()
	}

	fmt.Println("\n=== RSS Feed Test Completed ===")
}
