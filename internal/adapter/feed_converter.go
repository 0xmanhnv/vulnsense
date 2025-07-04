package adapter

import (
	"context"
	"vulnsense/internal/domain"
	"vulnsense/pkg/feed"
)

// ConvertToDomainVulnerability converts pkg/feed.Vulnerability to internal/domain.Vulnerability
func ConvertToDomainVulnerability(feedVuln feed.Vulnerability) domain.Vulnerability {
	// Convert CVSS if available
	var cvss domain.CVSS
	if feedVuln.CVSS != nil {
		cvss = domain.CVSS{
			V2Score:  feedVuln.CVSS.V2Score,
			V3Score:  feedVuln.CVSS.V3Score,
			V2Vector: feedVuln.CVSS.V2Vector,
			V3Vector: feedVuln.CVSS.V3Vector,
			Severity: convertSeverity(feedVuln.Severity),
		}
	} else {
		cvss = domain.CVSS{
			Severity: convertSeverity(feedVuln.Severity),
		}
	}

	// Create product
	product := domain.Product{
		Name:    feedVuln.ProductName,
		Version: feedVuln.ProductVersion,
	}

	return domain.Vulnerability{
		ID:               feedVuln.ID,
		Source:           feedVuln.Source,
		Product:          product,
		Title:            feedVuln.Title,
		Description:      feedVuln.Description,
		CVSS:             cvss,
		Severity:         convertSeverity(feedVuln.Severity),
		PublishedDate:    feedVuln.PublishedDate,
		LastModifiedDate: feedVuln.LastModifiedDate,
		References:       feedVuln.References,
		ExploitAvailable: feedVuln.ExploitAvailable,
		Aliases:          feedVuln.CVEIDs,
		Tags:             feedVuln.Tags,
	}
}

// convertSeverity converts string severity to domain.Severity
func convertSeverity(severity string) domain.Severity {
	switch severity {
	case "CRITICAL":
		return domain.SeverityCritical
	case "HIGH":
		return domain.SeverityHigh
	case "MEDIUM":
		return domain.SeverityMedium
	case "LOW":
		return domain.SeverityLow
	case "NONE":
		return domain.SeverityNone
	default:
		return domain.SeverityUnknown
	}
}

// FeedProviderAdapter adapts pkg/feed.Fetcher to work with domain types
type FeedProviderAdapter struct {
	fetcher feed.Fetcher
}

// NewFeedProviderAdapter creates a new adapter
func NewFeedProviderAdapter(fetcher feed.Fetcher) *FeedProviderAdapter {
	return &FeedProviderAdapter{
		fetcher: fetcher,
	}
}

// Fetch implements usecase.FeedProvider interface
func (f *FeedProviderAdapter) Fetch(ctx context.Context) ([]domain.Vulnerability, error) {
	feedVulns, err := f.fetcher.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	domainVulns := make([]domain.Vulnerability, len(feedVulns))
	for i, feedVuln := range feedVulns {
		domainVulns[i] = ConvertToDomainVulnerability(feedVuln)
	}

	return domainVulns, nil
}
