package feed

import "context"

// Fetcher defines the interface for fetching vulnerability data from feeds
type Fetcher interface {
	Fetch(ctx context.Context) ([]Vulnerability, error)
	SourceName() string
}
