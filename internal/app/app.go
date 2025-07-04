package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"vulnsense/internal/adapter"
	"vulnsense/internal/config"
	"vulnsense/internal/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
)

// App holds the application's dependencies and executes the main logic.
type App struct {
	logger         *slog.Logger
	fetchVulns     *usecase.FetchVulnerabilitiesUseCase
	fetchAssets    *usecase.FetchAssetsUseCase
	matchAndRecord *usecase.MatchAndRecordUseCase
	notify         *usecase.NotifyNewFindingsUseCase
	dbPool         *pgxpool.Pool
}

// FetchVulnsUseCase returns the FetchVulnerabilitiesUseCase.
func (a *App) FetchVulnsUseCase() *usecase.FetchVulnerabilitiesUseCase {
	return a.fetchVulns
}

// New creates and initializes a new App instance.
// It handles configuration loading, logger setup, database connection, and dependency injection.
func New(ctx context.Context) (*App, error) {
	// 1. Initialize Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Initializing application")

	// 2. Load Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// 3. Initialize DB Connection
	dbCtx, dbCancel := context.WithTimeout(ctx, 10*time.Second)
	defer dbCancel()

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.SSLMode)

	pool, err := pgxpool.New(dbCtx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	// No need to defer pool.Close() here, as the pool's lifetime is tied to the app.
	// It should be closed when the app shuts down. We'll handle that in main.

	if err := pool.Ping(dbCtx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	logger.Info("Database connection successful")

	app := &App{
		logger: logger,
		dbPool: pool,
	}

	// -- Repositories --
	vulnRepo, err := adapter.NewPostgresVulnerabilityRepository(pool)
	if err != nil {
		return nil, fmt.Errorf("failed to create vulnerability repository: %w", err)
	}

	assetRepo, err := adapter.NewPostgresAssetRepository(pool)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset repository: %w", err)
	}

	findingRepo, err := adapter.NewPostgresFindingRepository(pool)
	if err != nil {
		return nil, fmt.Errorf("failed to create finding repository: %w", err)
	}

	// -- Adapters --
	alerter := adapter.NewTelegramAlerter(cfg.Telegram, logger)
	assetFetcher := adapter.NewSplunkAssetFetcher(cfg.Splunk, logger)
	feedProviders := adapter.NewFeedProviders(cfg, logger)

	// -- Use Cases --
	app.fetchVulns = usecase.NewFetchVulnerabilitiesUseCase(feedProviders, vulnRepo, logger)
	app.fetchAssets = usecase.NewFetchAssetsUseCase(assetFetcher, assetRepo, logger)
	app.matchAndRecord = usecase.NewMatchAndRecordUseCase(vulnRepo, assetRepo, findingRepo, logger)
	app.notify = usecase.NewNotifyNewFindingsUseCase(findingRepo, vulnRepo, assetRepo, alerter, logger)

	return app, nil
}

// Run starts the application's main logic.
func (a *App) Run(ctx context.Context) error {
	a.logger.Info("Application starting")

	if err := a.fetchVulns.Execute(ctx); err != nil {
		a.logger.Error("Failed to fetch vulnerabilities", "error", err)
		// Depending on the policy, we might want to return here.
		// For now, we continue.
	}

	if err := a.fetchAssets.Execute(ctx); err != nil {
		a.logger.Error("Failed to fetch assets", "error", err)
	}

	if err := a.matchAndRecord.Execute(ctx); err != nil {
		a.logger.Error("Failed during matching and recording", "error", err)
	}

	if err := a.notify.Execute(ctx); err != nil {
		a.logger.Error("Failed during notification", "error", err)
	}

	a.logger.Info("Application run finished")
	return nil
}

// Stop gracefully shuts down the application.
func (a *App) Stop() {
	a.logger.Info("Shutting down application")
	if a.dbPool != nil {
		a.dbPool.Close()
		a.logger.Info("Database connection pool closed")
	}
}

// END OF FILE
