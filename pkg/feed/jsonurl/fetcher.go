package jsonurl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"vulnsense/internal/domain"
)

// Fetcher implements the feed.Fetcher interface for a generic JSON data source from a URL.
// It expects the JSON content at the URL to be an array of objects that can be
// unmarshalled into the []domain.Vulnerability struct.
type Fetcher struct {
	name   string
	url    string
	client *http.Client
	logger *slog.Logger
}

// New creates a new Fetcher for a given JSON feed URL.
// The 'name' is a human-readable identifier from config, e.g., "Custom CVE List".
func New(name, url string, client *http.Client, logger *slog.Logger) *Fetcher {
	return &Fetcher{
		name:   name,
		url:    url,
		client: client,
		logger: logger.With("feed_name", name, "feed_type", "jsonurl"),
	}
}

// SourceName returns the configured name of the feed.
func (f *Fetcher) SourceName() string {
	return f.name
}

// Fetch retrieves and parses vulnerability data from the configured URL.
func (f *Fetcher) Fetch(ctx context.Context) ([]domain.Vulnerability, error) {
	f.logger.Info("Starting fetch from json url feed", "url", f.url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for url %s: %w", f.url, err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request for url %s: %w", f.url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code from %s: %d", f.url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", f.url, err)
	}

	var vulnerabilities []domain.Vulnerability
	if err := json.Unmarshal(body, &vulnerabilities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json from %s: %w", f.url, err)
	}

	f.logger.Info("Successfully fetched vulnerabilities", "count", len(vulnerabilities))
	return vulnerabilities, nil
}
