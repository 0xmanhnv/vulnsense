package usecase

import (
	"context"
	"fmt"
	"log/slog"
)

// FetchAssetsUseCase is responsible for fetching assets from an external source
// and persisting them in the local database.
type FetchAssetsUseCase struct {
	fetcher    AssetFetcher
	repository AssetRepository
	logger     *slog.Logger
}

// NewFetchAssetsUseCase creates a new instance of FetchAssetsUseCase.
func NewFetchAssetsUseCase(
	fetcher AssetFetcher,
	repository AssetRepository,
	logger *slog.Logger,
) *FetchAssetsUseCase {
	return &FetchAssetsUseCase{
		fetcher:    fetcher,
		repository: repository,
		logger:     logger,
	}
}

// Execute runs the use case.
func (uc *FetchAssetsUseCase) Execute(ctx context.Context) error {
	uc.logger.Info("Starting asset fetching process from external source")

	assets, err := uc.fetcher.FetchAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch assets from external source: %w", err)
	}

	uc.logger.Info("Successfully fetched assets from external source", "count", len(assets))

	var savedCount int
	for _, asset := range assets {
		if err := uc.repository.Save(ctx, asset); err != nil {
			uc.logger.Error("Failed to save asset to the database", "asset_id", asset.ID, "asset_name", asset.Name, "error", err)
			// Continue to the next asset, do not stop the whole process
		} else {
			savedCount++
		}
	}

	uc.logger.Info("Asset fetching process finished", "total_fetched", len(assets), "successfully_saved", savedCount)
	return nil
}
