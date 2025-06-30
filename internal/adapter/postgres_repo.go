package adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"vulnsense/internal/domain"
	"vulnsense/internal/usecase"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresVulnerabilityRepository is the PostgreSQL implementation of the VulnerabilityRepository.
type PostgresVulnerabilityRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresVulnerabilityRepository creates a new instance of PostgresVulnerabilityRepository.
// It takes a database connection pool and returns a repository that satisfies the usecase.VulnerabilityRepository interface.
func NewPostgresVulnerabilityRepository(pool *pgxpool.Pool) (usecase.VulnerabilityRepository, error) {
	if pool == nil {
		return nil, fmt.Errorf("database connection pool cannot be nil")
	}
	return &PostgresVulnerabilityRepository{pool: pool}, nil
}

// Save inserts or updates a vulnerability record in the database.
// It uses an "upsert" operation (ON CONFLICT DO UPDATE) to handle both cases.
func (r *PostgresVulnerabilityRepository) Save(ctx context.Context, v domain.Vulnerability) error {
	query := `
		INSERT INTO vulnerabilities (
			id, source, description, published_date, last_modified_date, exploit_available, 
			product, cvss, "references", title, severity, affected_versions, fixed_versions, 
			aliases, tags, cwe, epss
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (id) DO UPDATE SET
			source = EXCLUDED.source,
			description = EXCLUDED.description,
			published_date = EXCLUDED.published_date,
			last_modified_date = EXCLUDED.last_modified_date,
			exploit_available = EXCLUDED.exploit_available,
			product = EXCLUDED.product,
			cvss = EXCLUDED.cvss,
			"references" = EXCLUDED."references",
			title = EXCLUDED.title,
			severity = EXCLUDED.severity,
			affected_versions = EXCLUDED.affected_versions,
			fixed_versions = EXCLUDED.fixed_versions,
			aliases = EXCLUDED.aliases,
			tags = EXCLUDED.tags,
			cwe = EXCLUDED.cwe,
			epss = EXCLUDED.epss;
	`
	_, err := r.pool.Exec(ctx, query,
		v.ID, v.Source, v.Description, v.PublishedDate, v.LastModifiedDate, v.ExploitAvailable,
		v.Product, v.CVSS, v.References, v.Title, v.Severity, v.AffectedVersions, v.FixedVersions,
		v.Aliases, v.Tags, v.CWE, v.EPSS,
	)
	return err
}

// FindByID retrieves a vulnerability by its ID from the database.
func (r *PostgresVulnerabilityRepository) FindByID(ctx context.Context, id string) (domain.Vulnerability, error) {
	var v domain.Vulnerability
	query := `
		SELECT 
			id, source, description, published_date, last_modified_date, exploit_available, 
			product, cvss, "references", title, severity, affected_versions, fixed_versions, 
			aliases, tags, cwe, epss
		FROM vulnerabilities
		WHERE id = $1;
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&v.ID, &v.Source, &v.Description, &v.PublishedDate, &v.LastModifiedDate, &v.ExploitAvailable,
		&v.Product, &v.CVSS, &v.References, &v.Title, &v.Severity, &v.AffectedVersions, &v.FixedVersions,
		&v.Aliases, &v.Tags, &v.CWE, &v.EPSS,
	)
	if err != nil {
		return domain.Vulnerability{}, err
	}
	return v, nil
}

// FindAll retrieves all vulnerabilities from the database.
// In a real-world application, this should be paginated.
func (r *PostgresVulnerabilityRepository) FindAll(ctx context.Context) ([]domain.Vulnerability, error) {
	var vulnerabilities []domain.Vulnerability
	query := `
		SELECT 
			id, source, description, published_date, last_modified_date, exploit_available, 
			product, cvss, "references", title, severity, affected_versions, fixed_versions, 
			aliases, tags, cwe, epss
		FROM vulnerabilities;
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var v domain.Vulnerability
		if err := rows.Scan(
			&v.ID, &v.Source, &v.Description, &v.PublishedDate, &v.LastModifiedDate, &v.ExploitAvailable,
			&v.Product, &v.CVSS, &v.References, &v.Title, &v.Severity, &v.AffectedVersions, &v.FixedVersions,
			&v.Aliases, &v.Tags, &v.CWE, &v.EPSS,
		); err != nil {
			return nil, err
		}
		vulnerabilities = append(vulnerabilities, v)
	}

	return vulnerabilities, nil
}

// --- Finding Repository Implementation ---

// PostgresFindingRepository is the PostgreSQL implementation of the FindingRepository.
type PostgresFindingRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresFindingRepository creates a new instance of PostgresFindingRepository.
func NewPostgresFindingRepository(pool *pgxpool.Pool) (usecase.FindingRepository, error) {
	if pool == nil {
		return nil, fmt.Errorf("database connection pool cannot be nil")
	}
	return &PostgresFindingRepository{pool: pool}, nil
}

// Save inserts or updates a finding record.
// It uses ON CONFLICT on the unique constraint (asset_id, vulnerability_id) to update existing records.
func (r *PostgresFindingRepository) Save(ctx context.Context, f domain.Finding) error {
	query := `
		INSERT INTO findings (id, asset_id, vulnerability_id, detected_at, status, patched_at, scanner, detection_method, package_path, evidence)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (asset_id, vulnerability_id) DO UPDATE SET
			status = EXCLUDED.status,
			detected_at = EXCLUDED.detected_at, -- Or maybe keep the first time? Depends on business logic.
			patched_at = EXCLUDED.patched_at,
			scanner = EXCLUDED.scanner,
			detection_method = EXCLUDED.detection_method,
			package_path = EXCLUDED.package_path,
			evidence = EXCLUDED.evidence;
	`
	_, err := r.pool.Exec(ctx, query,
		f.ID, f.AssetID, f.VulnerabilityID, f.DetectedAt, f.Status, f.PatchedAt,
		f.Scanner, f.DetectionMethod, f.PackagePath, f.Evidence,
	)
	return err
}

// GetNewFindings retrieves all findings that are currently in the 'OPEN' state.
// This is a simplified implementation for the notification use case.
func (r *PostgresFindingRepository) GetNewFindings(ctx context.Context) ([]domain.Finding, error) {
	var findings []domain.Finding
	query := `
		SELECT id, asset_id, vulnerability_id, detected_at, status, patched_at, scanner, detection_method, package_path, evidence
		FROM findings
		WHERE status = $1;
	`
	rows, err := r.pool.Query(ctx, query, domain.FindingStatusOpen)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var f domain.Finding
		if err := rows.Scan(
			&f.ID, &f.AssetID, &f.VulnerabilityID, &f.DetectedAt, &f.Status, &f.PatchedAt,
			&f.Scanner, &f.DetectionMethod, &f.PackagePath, &f.Evidence,
		); err != nil {
			return nil, err
		}
		findings = append(findings, f)
	}
	return findings, nil
}

// --- Asset Repository Implementation ---

// PostgresAssetRepository is the PostgreSQL implementation of the AssetRepository.
type PostgresAssetRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresAssetRepository creates a new instance of PostgresAssetRepository.
func NewPostgresAssetRepository(pool *pgxpool.Pool) (usecase.AssetRepository, error) {
	if pool == nil {
		return nil, fmt.Errorf("database connection pool cannot be nil")
	}
	return &PostgresAssetRepository{pool: pool}, nil
}

// scanAsset is a helper function to scan a row from the database and convert it into a domain.Asset struct.
// It handles the polymorphic 'details' field.
func scanAsset(row pgx.Row) (domain.Asset, error) {
	var a domain.Asset
	var detailsJSON, tagsJSON, installedProductsJSON, metadataJSON []byte

	err := row.Scan(
		&a.ID, &a.Name, &a.Type, &a.OwnerTeam, &a.FirstSeen, &a.LastSeen,
		&detailsJSON, &installedProductsJSON, &tagsJSON, &metadataJSON,
	)
	if err != nil {
		return domain.Asset{}, err
	}

	// Unmarshal generic JSON fields
	if tagsJSON != nil {
		if err := json.Unmarshal(tagsJSON, &a.Tags); err != nil {
			return domain.Asset{}, fmt.Errorf("failed to unmarshal asset tags: %w", err)
		}
	}
	if installedProductsJSON != nil {
		if err := json.Unmarshal(installedProductsJSON, &a.InstalledProducts); err != nil {
			return domain.Asset{}, fmt.Errorf("failed to unmarshal asset installed products: %w", err)
		}
	}
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &a.Metadata); err != nil {
			return domain.Asset{}, fmt.Errorf("failed to unmarshal asset metadata: %w", err)
		}
	}

	// Unmarshal the polymorphic 'details' field based on the asset type
	if detailsJSON != nil {
		switch a.Type {
		case domain.AssetTypeHost:
			var details domain.HostDetails
			if err := json.Unmarshal(detailsJSON, &details); err != nil {
				return domain.Asset{}, fmt.Errorf("failed to unmarshal host details: %w", err)
			}
			a.Details = details
		case domain.AssetTypeGitRepository:
			var details domain.GitRepositoryDetails
			if err := json.Unmarshal(detailsJSON, &details); err != nil {
				return domain.Asset{}, fmt.Errorf("failed to unmarshal git repo details: %w", err)
			}
			a.Details = details
		case domain.AssetTypeContainerImage:
			var details domain.ContainerImageDetails
			if err := json.Unmarshal(detailsJSON, &details); err != nil {
				return domain.Asset{}, fmt.Errorf("failed to unmarshal container image details: %w", err)
			}
			a.Details = details
		case domain.AssetTypeWebApplication:
			var details domain.WebApplicationDetails
			if err := json.Unmarshal(detailsJSON, &details); err != nil {
				return domain.Asset{}, fmt.Errorf("failed to unmarshal web app details: %w", err)
			}
			a.Details = details
		}
	}

	return a, nil
}

// FindByID retrieves an asset by its ID from the database.
func (r *PostgresAssetRepository) FindByID(ctx context.Context, id string) (domain.Asset, error) {
	query := `
		SELECT id, name, type, owner_team, first_seen, last_seen,
		       details, installed_products, tags, metadata
		FROM assets
		WHERE id = $1;
	`
	row := r.pool.QueryRow(ctx, query, id)
	return scanAsset(row)
}

// Save inserts or updates an asset record.
// It uses ON CONFLICT on the 'id' column to update existing records.
func (r *PostgresAssetRepository) Save(ctx context.Context, a domain.Asset) error {
	detailsJSON, err := json.Marshal(a.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal asset details: %w", err)
	}

	tagsJSON, err := json.Marshal(a.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal asset tags: %w", err)
	}

	query := `
		INSERT INTO assets (id, name, type, owner_team, first_seen, last_seen, details, installed_products, tags, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			owner_team = EXCLUDED.owner_team,
			last_seen = EXCLUDED.last_seen,
			details = EXCLUDED.details,
			installed_products = EXCLUDED.installed_products,
			tags = EXCLUDED.tags,
			metadata = EXCLUDED.metadata;
	`
	_, err = r.pool.Exec(ctx, query,
		a.ID, a.Name, a.Type, a.OwnerTeam, a.FirstSeen, a.LastSeen,
		detailsJSON, a.InstalledProducts, tagsJSON, a.Metadata,
	)
	return err
}

// FindAll retrieves all assets from the database.
func (r *PostgresAssetRepository) FindAll(ctx context.Context) ([]domain.Asset, error) {
	var assets []domain.Asset
	query := `
		SELECT id, name, type, owner_team, first_seen, last_seen,
		       details, installed_products, tags, metadata
		FROM assets;
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, nil
}
