package database

import (
	"context"
	"fmt"
	"time"

	"vulnsense/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQLProvider implements DatabaseProvider for PostgreSQL
type PostgreSQLProvider struct{}

func NewPostgreSQLProvider() DatabaseProvider {
	return &PostgreSQLProvider{}
}

func (p *PostgreSQLProvider) GetType() DatabaseType {
	return DatabaseTypePostgreSQL
}

func (p *PostgreSQLProvider) CreateDatabase(config config.DatabaseConfig) (Database, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.Name, config.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PostgreSQL config: %w", err)
	}

	// Configure connection pool
	if config.MaxConnections > 0 {
		poolConfig.MaxConns = int32(config.MaxConnections)
	}
	if config.MinConnections > 0 {
		poolConfig.MinConns = int32(config.MinConnections)
	}
	if config.MaxLifetime > 0 {
		poolConfig.MaxConnLifetime = config.MaxLifetime
	}
	if config.MaxIdleTime > 0 {
		poolConfig.MaxConnIdleTime = config.MaxIdleTime
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL pool: %w", err)
	}

	return &PostgreSQLDatabase{
		pool:   pool,
		config: config,
		dsn:    dsn,
	}, nil
}

func (p *PostgreSQLProvider) ValidateConfig(config config.DatabaseConfig) error {
	if config.Host == "" {
		return fmt.Errorf("host is required for PostgreSQL")
	}
	if config.Port <= 0 {
		return fmt.Errorf("valid port is required for PostgreSQL")
	}
	if config.Name == "" {
		return fmt.Errorf("database name is required for PostgreSQL")
	}
	if config.User == "" {
		return fmt.Errorf("user is required for PostgreSQL")
	}
	return nil
}

func (p *PostgreSQLProvider) GetDefaultConfig() config.DatabaseConfig {
	return config.DatabaseConfig{
		Type:           string(DatabaseTypePostgreSQL),
		Host:           "localhost",
		Port:           5432,
		SSLMode:        "require",
		MaxConnections: 10,
		MinConnections: 2,
		MaxLifetime:    time.Hour,
		MaxIdleTime:    time.Minute * 30,
	}
}

// PostgreSQLDatabase implements Database interface for PostgreSQL
type PostgreSQLDatabase struct {
	pool   *pgxpool.Pool
	config config.DatabaseConfig
	dsn    string
}

func (db *PostgreSQLDatabase) Connect(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

func (db *PostgreSQLDatabase) Close() error {
	if db.pool != nil {
		db.pool.Close()
	}
	return nil
}

func (db *PostgreSQLDatabase) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

func (db *PostgreSQLDatabase) HealthCheck(ctx context.Context) *HealthStatus {
	start := time.Now()

	if err := db.Ping(ctx); err != nil {
		return &HealthStatus{
			Status:       "unhealthy",
			ResponseTime: time.Since(start),
			Message:      fmt.Sprintf("ping failed: %v", err),
		}
	}

	stats := db.pool.Stat()
	return &HealthStatus{
		Status:          "healthy",
		ResponseTime:    time.Since(start),
		OpenConnections: int32(stats.TotalConns()),
		IdleConnections: int32(stats.IdleConns()),
	}
}

func (db *PostgreSQLDatabase) Stats() DatabaseStats {
	stats := db.pool.Stat()
	return DatabaseStats{
		OpenConnections:     int32(stats.TotalConns()),
		IdleConnections:     int32(stats.IdleConns()),
		AcquiredConnections: int32(stats.AcquiredConns()),
		MaxConnections:      int32(stats.MaxConns()),
	}
}

func (db *PostgreSQLDatabase) WithTransaction(ctx context.Context, fn TransactionFunc) error {
	return db.WithTransactionRetry(ctx, 1, fn)
}

func (db *PostgreSQLDatabase) WithTransactionRetry(ctx context.Context, maxRetries int, fn TransactionFunc) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = db.executeTransaction(ctx, fn)
		if err == nil {
			return nil
		}

		// Check if error is retryable
		if !isRetryableError(err) {
			return err
		}

		// Wait before retry
		if i < maxRetries-1 {
			time.Sleep(time.Millisecond * 100 * time.Duration(i+1))
		}
	}
	return err
}

func (db *PostgreSQLDatabase) executeTransaction(ctx context.Context, fn TransactionFunc) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	pgxTx := &PostgreSQLTransaction{tx: tx}
	err = fn(pgxTx)
	return err
}

func (db *PostgreSQLDatabase) GetConnection() interface{} {
	return db.pool
}

func (db *PostgreSQLDatabase) GetType() DatabaseType {
	return DatabaseTypePostgreSQL
}

func (db *PostgreSQLDatabase) GetConnectionString() string {
	return db.dsn
}

func (db *PostgreSQLDatabase) GetMigrator() Migrator {
	return NewPostgreSQLMigrator(db)
}

// PostgreSQLTransaction implements Transaction interface
type PostgreSQLTransaction struct {
	tx pgx.Tx
}

func (t *PostgreSQLTransaction) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	result, err := t.tx.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PostgreSQLResult{rowsAffected: result.RowsAffected()}, nil
}

func (t *PostgreSQLTransaction) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := t.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PostgreSQLRows{rows: rows}, nil
}

func (t *PostgreSQLTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	row := t.tx.QueryRow(ctx, query, args...)
	return &PostgreSQLRow{row: row}
}

func (t *PostgreSQLTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *PostgreSQLTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// PostgreSQL-specific implementations of Result, Rows, Row
type PostgreSQLResult struct {
	rowsAffected int64
}

func (r *PostgreSQLResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

func (r *PostgreSQLResult) LastInsertId() (int64, error) {
	// PostgreSQL doesn't have LastInsertId concept like MySQL
	return 0, fmt.Errorf("PostgreSQL does not support LastInsertId, use RETURNING clause instead")
}

type PostgreSQLRows struct {
	rows pgx.Rows
}

func (r *PostgreSQLRows) Next() bool {
	return r.rows.Next()
}

func (r *PostgreSQLRows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

func (r *PostgreSQLRows) Close() error {
	r.rows.Close()
	return nil
}

func (r *PostgreSQLRows) Err() error {
	return r.rows.Err()
}

type PostgreSQLRow struct {
	row pgx.Row
}

func (r *PostgreSQLRow) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}

// PostgreSQL migrator placeholder
type PostgreSQLMigrator struct {
	db *PostgreSQLDatabase
}

func NewPostgreSQLMigrator(db *PostgreSQLDatabase) Migrator {
	return &PostgreSQLMigrator{db: db}
}

func (m *PostgreSQLMigrator) RunMigrations(ctx context.Context, migrationsPath string) error {
	// TODO: Implement migration logic
	return fmt.Errorf("PostgreSQL migrations not implemented yet")
}

func (m *PostgreSQLMigrator) GetVersion(ctx context.Context) (int, error) {
	// TODO: Implement version tracking
	return 0, fmt.Errorf("PostgreSQL version tracking not implemented yet")
}

// Helper function to check if error is retryable
func isRetryableError(err error) bool {
	// Add logic to check if the error is retryable (e.g., connection issues)
	return false
}
