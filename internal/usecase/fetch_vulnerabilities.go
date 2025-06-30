package usecase

import (
	"context"
	"log/slog"
)

// FetchVulnerabilitiesUseCase is the use case for fetching vulnerabilities from various providers.
// It orchestrates the process of getting data from feeds and saving it to a repository.
type FetchVulnerabilitiesUseCase struct {
	providers []FeedProvider
	repo      VulnerabilityRepository
	logger    *slog.Logger
}

// NewFetchVulnerabilitiesUseCase creates a new instance of FetchVulnerabilitiesUseCase.
func NewFetchVulnerabilitiesUseCase(providers []FeedProvider, repo VulnerabilityRepository, logger *slog.Logger) *FetchVulnerabilitiesUseCase {
	return &FetchVulnerabilitiesUseCase{
		providers: providers,
		repo:      repo,
		logger:    logger,
	}
}

// Execute runs the use case.
// It iterates through all configured feed providers, fetches the vulnerabilities from each,
// and saves them to the repository.
func (uc *FetchVulnerabilitiesUseCase) Execute(ctx context.Context) error {
	uc.logger.Info("Starting vulnerability fetch process")

	for _, provider := range uc.providers {
		vulnerabilities, err := provider.Fetch(ctx)
		if err != nil {
			uc.logger.Error("Failed to fetch vulnerabilities from a provider", "error", err)
			// Decide if we should continue with other providers or return an error.
			// For now, we'll log the error and continue.
			continue
		}

		uc.logger.Info("Fetched vulnerabilities", "count", len(vulnerabilities))

		for _, vuln := range vulnerabilities {
			if err := uc.repo.Save(ctx, vuln); err != nil {
				uc.logger.Error("Failed to save vulnerability", "vulnerability_id", vuln.ID, "error", err)
				// Decide if we should continue. For now, we continue.
			}
		}
	}

	uc.logger.Info("Vulnerability fetch process finished")
	return nil
}
