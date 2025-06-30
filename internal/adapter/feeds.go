package adapter

import (
	"context"
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
			AffectedVersions: []string{"<=2.15.0"},
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
			AffectedVersions: []string{"5.3.0 to 5.3.17", "5.2.0 to 5.2.19"},
		},
	}, nil
}

func NewFeedProviders() []usecase.FeedProvider {
	return []usecase.FeedProvider{
		&OpenCVEProvider{},
		&NVDProvider{},
	}
}
