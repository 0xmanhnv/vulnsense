package feed

import (
	"context"
	"vulnsense/internal/domain"
)

// Fetcher defines the standard interface for any vulnerability data source.
// It decouples the application's core logic from the specific implementation
// details of how data is retrieved from a particular source.
type Fetcher interface {
	// Fetch retrieves vulnerability data from the source.
	Fetch(ctx context.Context) ([]domain.Vulnerability, error)

	// SourceName returns a unique, human-readable identifier for the feed.
	// e.g., "OpenCVE", "NVD-API-v2".
	SourceName() string
}
