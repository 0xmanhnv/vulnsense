package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"vulnsense/internal/domain"
)

// NotifyNewFindingsUseCase is responsible for fetching new findings and sending alerts.
type NotifyNewFindingsUseCase struct {
	findingRepo FindingRepository
	vulnRepo    VulnerabilityRepository // Needed to get vuln details for the alert
	assetRepo   AssetRepository         // Needed to get asset details for the alert
	alerter     Alerter
	logger      *slog.Logger
}

// NewNotifyNewFindingsUseCase creates a new instance of NotifyNewFindingsUseCase.
func NewNotifyNewFindingsUseCase(
	findingRepo FindingRepository,
	vulnRepo VulnerabilityRepository,
	assetRepo AssetRepository,
	alerter Alerter,
	logger *slog.Logger,
) *NotifyNewFindingsUseCase {
	return &NotifyNewFindingsUseCase{
		findingRepo: findingRepo,
		vulnRepo:    vulnRepo,
		assetRepo:   assetRepo,
		alerter:     alerter,
		logger:      logger,
	}
}

// Execute runs the use case.
func (uc *NotifyNewFindingsUseCase) Execute(ctx context.Context) error {
	uc.logger.Info("Checking for new findings to notify")

	newFindings, err := uc.findingRepo.GetNewFindings(ctx)
	if err != nil {
		return fmt.Errorf("could not get new findings: %w", err)
	}

	if len(newFindings) == 0 {
		uc.logger.Info("No new findings to notify")
		return nil
	}

	uc.logger.Info("Found new findings, preparing to send alerts", "count", len(newFindings))

	for _, finding := range newFindings {
		vuln, err := uc.vulnRepo.FindByID(ctx, finding.VulnerabilityID)
		if err != nil {
			uc.logger.Error("Failed to get vulnerability details for finding", "finding_id", finding.ID, "vuln_id", finding.VulnerabilityID, "error", err)
			continue
		}

		asset, err := uc.assetRepo.FindByID(ctx, finding.AssetID)
		if err != nil {
			uc.logger.Error("Failed to get asset details for finding", "finding_id", finding.ID, "asset_id", finding.AssetID, "error", err)
			continue
		}

		message := uc.formatAlertMessage(asset, vuln)

		if err := uc.alerter.Notify(ctx, message); err != nil {
			uc.logger.Error("Failed to send alert for finding", "finding_id", finding.ID, "error", err)
			// Continue to next finding
		} else {
			// TODO: Here you would update the finding to mark it as "notified"
			uc.logger.Info("Successfully sent alert for finding", "finding_id", finding.ID)
		}
	}
	return nil
}

// formatAlertMessage creates a detailed alert message based on the asset's type.
func (uc *NotifyNewFindingsUseCase) formatAlertMessage(asset domain.Asset, vuln domain.Vulnerability) string {
	var assetDetails string

	// Use a type switch to handle different kinds of assets.
	switch details := asset.Details.(type) {
	case domain.HostDetails:
		assetDetails = fmt.Sprintf("*Asset*: %s (Type: %s, IP: `%s`)", asset.Name, asset.Type, details.IPAddress)
	case domain.GitRepositoryDetails:
		assetDetails = fmt.Sprintf("*Asset*: %s (Type: %s, URL: `%s`)", asset.Name, asset.Type, details.URL)
	case domain.ContainerImageDetails:
		assetDetails = fmt.Sprintf("*Asset*: %s (Type: %s, Image: `%s:%s`)", asset.Name, asset.Type, details.Name, details.Tag)
	default:
		assetDetails = fmt.Sprintf("*Asset*: %s (Type: %s)", asset.Name, asset.Type)
	}

	return fmt.Sprintf(
		"*Vulnerability Alert* `[%s]`\n\n"+
			"*Title*: %s\n"+
			"%s", // Placeholder for the asset details string
		vuln.Severity,
		vuln.Title,
		assetDetails,
	)
}
