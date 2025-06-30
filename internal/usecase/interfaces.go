package usecase

import (
	"context"

	"vulnsense/internal/domain"
)

// VulnerabilityRepository defines the interface for storing and retrieving vulnerability data.
// It acts as a port to the data persistence layer, abstracting the database details from the use cases.
type VulnerabilityRepository interface {
	Save(ctx context.Context, vulnerability domain.Vulnerability) error
	FindByID(ctx context.Context, id string) (domain.Vulnerability, error)
	FindAll(ctx context.Context) ([]domain.Vulnerability, error)
	// We can add more methods like FindByCVSS, FindBySeverity, etc. as needed.
}

// AssetFetcher represents an external source for asset data, like Splunk or an API.
// It is now a distinct interface from the repository.
type AssetFetcher interface {
	FetchAll(ctx context.Context) ([]domain.Asset, error)
}

// Alerter defines the interface for sending notifications.
// This port abstracts the notification mechanism, allowing use cases to send alerts without being coupled to a specific service like Slack or Email.
type Alerter interface {
	Notify(ctx context.Context, message string) error
}

// FeedProvider defines the interface for a source of vulnerability data.
// Each feed (e.g., NVD, OpenCVE) will have an adapter that implements this interface.
type FeedProvider interface {
	Fetch(ctx context.Context) ([]domain.Vulnerability, error)
}

// FindingRepository defines the interface for storing and retrieving finding data.
// A "Finding" is a specific instance of a vulnerability on a specific asset.
type FindingRepository interface {
	Save(ctx context.Context, finding domain.Finding) error
	// This will be used by the notification use case to get new findings.
	GetNewFindings(ctx context.Context) ([]domain.Finding, error)
}

// AssetRepository defines the interface for retrieving asset data.
// It will likely have more methods in the future.
type AssetRepository interface {
	FindByID(ctx context.Context, id string) (domain.Asset, error)
	FindAll(ctx context.Context) ([]domain.Asset, error)
	Save(ctx context.Context, asset domain.Asset) error
}
