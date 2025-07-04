package adapter

import (
	"context"
	"log/slog"
	"net/http"
	"time"
	"vulnsense/internal/config"
	"vulnsense/internal/domain"
	"vulnsense/internal/usecase"
)

// OpenCVEProvider is a placeholder for a real OpenCVE feed provider.
type OpenCVEProvider struct{}

// Fetch retrieves vulnerabilities from the (placeholder) OpenCVE source.
func (p *OpenCVEProvider) Fetch(ctx context.Context) ([]domain.Vulnerability, error) {
	// In a real implementation, you would make an HTTP request to the OpenCVE API.
	return []domain.Vulnerability{
		{
			ID:          "CVE-2021-44228",
			Title:       "Apache Log4j2 JNDI features do not protect against attacker controlled LDAP and other JNDI related endpoints",
			Severity:    "CRITICAL",
			Description: "A critical vulnerability in Apache Log4j2...",
			Product: domain.Product{
				Name: "Log4j",
			},
		},
	}, nil
}

// NVDProvider is a placeholder for a real NVD feed provider.
type NVDProvider struct{}

// Fetch retrieves vulnerabilities from the (placeholder) NVD source.
func (p *NVDProvider) Fetch(ctx context.Context) ([]domain.Vulnerability, error) {
	// In a real implementation, you would make an HTTP request to the NVD API.
	return []domain.Vulnerability{
		{
			ID:          "CVE-2022-22965",
			Title:       "Spring Framework RCE via Data Binding on JDK 9+",
			Severity:    "CRITICAL",
			Description: "A Spring MVC or Spring WebFlux application running on JDK 9+ may be vulnerable to remote code execution (RCE) via data binding...",
			Product: domain.Product{
				Name: "Spring Framework",
			},
		},
	}, nil
}

// NewFeedProviders creates and returns all configured feed providers including RSS feeds
func NewFeedProviders(cfg *config.Config, logger *slog.Logger) []usecase.FeedProvider {
	providers := []usecase.FeedProvider{
		&OpenCVEProvider{},
		&NVDProvider{},
	}

	// Add RSS feed providers from configuration
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	for _, source := range cfg.RSSSources {
		if !source.Enabled {
			continue // Skip disabled feeds
		}

		// Create RSS adapter
		rssAdapter := NewRSSFeedAdapter(
			source.Name,
			source.URL,
			source.Type,
			source.Source,
			httpClient,
			logger,
		)

		// Wrap with converter adapter
		feedProvider := NewFeedProviderAdapter(rssAdapter)
		providers = append(providers, feedProvider)
	}

	logger.Info("Initialized feed providers", "count", len(providers), "rss_feeds", len(cfg.RSSSources))
	return providers
}
