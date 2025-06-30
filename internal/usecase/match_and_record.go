package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"vulnsense/internal/domain"

	"github.com/google/uuid"
)

// MatchAndRecordUseCase is responsible for matching vulnerabilities against assets and creating Finding records.
type MatchAndRecordUseCase struct {
	vulnRepo    VulnerabilityRepository
	assetRepo   AssetRepository
	findingRepo FindingRepository
	logger      *slog.Logger
}

// NewMatchAndRecordUseCase creates a new instance of MatchAndRecordUseCase.
func NewMatchAndRecordUseCase(
	vulnRepo VulnerabilityRepository,
	assetRepo AssetRepository,
	findingRepo FindingRepository,
	logger *slog.Logger,
) *MatchAndRecordUseCase {
	return &MatchAndRecordUseCase{
		vulnRepo:    vulnRepo,
		assetRepo:   assetRepo,
		findingRepo: findingRepo,
		logger:      logger,
	}
}

// Execute runs the use case.
func (uc *MatchAndRecordUseCase) Execute(ctx context.Context) error {
	uc.logger.Info("Starting asset matching and recording process")

	assets, err := uc.assetRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch assets from repository: %w", err)
	}
	uc.logger.Info("Fetched assets from repository", "count", len(assets))

	vulnerabilities, err := uc.vulnRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch vulnerabilities: %w", err)
	}
	uc.logger.Info("Fetched vulnerabilities", "count", len(vulnerabilities))

	for _, asset := range assets {
		// Only perform product-based matching for HOST assets.
		if asset.Type != domain.AssetTypeHost {
			continue
		}

		for _, installedProduct := range asset.InstalledProducts {
			for _, vuln := range vulnerabilities {
				if installedProduct.Name == vuln.Product.Name {
					uc.logger.Info("Found a potential match", "asset", asset.Name, "vulnerability", vuln.ID)

					finding := domain.Finding{
						ID:              uuid.NewString(), // Generate a new unique ID for the finding
						AssetID:         asset.ID,
						VulnerabilityID: vuln.ID,
						DetectedAt:      time.Now(),
						Status:          domain.FindingStatusOpen,
						Scanner:         "vulnsense-matcher",
						DetectionMethod: "product-name-match",
						PackagePath:     installedProduct.InstallPath,
					}

					if err := uc.findingRepo.Save(ctx, finding); err != nil {
						uc.logger.Error("Failed to save finding", "asset_id", asset.ID, "vuln_id", vuln.ID, "error", err)
						// Continue to the next match
					} else {
						uc.logger.Info("Successfully recorded new finding", "finding_id", finding.ID)
					}
				}
			}
		}
	}

	uc.logger.Info("Asset matching and recording process finished")
	return nil
}
