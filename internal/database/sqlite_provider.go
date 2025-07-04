package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"vulnsense/internal/config"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteProvider implements DatabaseProvider for SQLite
type SQLiteProvider struct{}

func NewSQLiteProvider() DatabaseProvider {
	return &SQLiteProvider{}
}

func (p *SQLiteProvider) GetType() DatabaseType {
	return DatabaseTypeSQLite
}

func (p *SQLiteProvider) CreateDatabase(config config.DatabaseConfig) (Database, error) {
	dsn := config.Name // For SQLite, Name is the file path
	if dsn == "" {
		dsn = ":memory:" // In-memory database
	}

	db, err := sql.Open("sqlite3", dsn+"?_journal_mode=WAL&_foreign_keys=1&_cache_size=-64000")
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite connection: %w", err)
	}

	// SQLite-specific settings
	db.SetMaxOpenConns(1) // SQLite doesn't support concurrent writes

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	return &SQLiteDatabase{
		db:     db,
		config: config,
		dsn:    dsn,
	}, nil
}

func (p *SQLiteProvider) ValidateConfig(config config.DatabaseConfig) error {
	// SQLite has minimal validation requirements
	if config.Name == "" {
		// Allow empty name for in-memory database
		return nil
	}
	return nil
}

func (p *SQLiteProvider) GetDefaultConfig() config.DatabaseConfig {
	return config.DatabaseConfig{
		Type:           string(DatabaseTypeSQLite),
		Name:           "vulnsense.db",
		MaxConnections: 1,
		MinConnections: 1,
	}
}

// SQLiteDatabase implements Database interface for SQLite
type SQLiteDatabase struct {
	db     *sql.DB
	config config.DatabaseConfig
	dsn    string
}

func (db *SQLiteDatabase) Connect(ctx context.Context) error {
	return db.db.PingContext(ctx)
}

func (db *SQLiteDatabase) Close() error {
	return db.db.Close()
}

func (db *SQLiteDatabase) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}

func (db *SQLiteDatabase) HealthCheck(ctx context.Context) *HealthStatus {
	start := time.Now()

	if err := db.Ping(ctx); err != nil {
		return &HealthStatus{
			Status:       "unhealthy",
			ResponseTime: time.Since(start),
			Message:      fmt.Sprintf("ping failed: %v", err),
		}
	}

	stats := db.db.Stats()
	return &HealthStatus{
		Status:          "healthy",
		ResponseTime:    time.Since(start),
		OpenConnections: int32(stats.OpenConnections),
		IdleConnections: int32(stats.Idle),
	}
}

func (db *SQLiteDatabase) Stats() DatabaseStats {
	stats := db.db.Stats()
	return DatabaseStats{
		OpenConnections:     int32(stats.OpenConnections),
		IdleConnections:     int32(stats.Idle),
		AcquiredConnections: int32(stats.InUse),
		MaxConnections:      int32(stats.MaxOpenConnections),
	}
}

func (db *SQLiteDatabase) WithTransaction(ctx context.Context, fn TransactionFunc) error {
	return db.WithTransactionRetry(ctx, 1, fn)
}

func (db *SQLiteDatabase) WithTransactionRetry(ctx context.Context, maxRetries int, fn TransactionFunc) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = db.executeTransaction(ctx, fn)
		if err == nil {
			return nil
		}

		// Check if error is retryable (e.g., SQLITE_BUSY)
		if !isSQLiteRetryableError(err) {
			return err
		}

		// Wait before retry
		if i < maxRetries-1 {
			time.Sleep(time.Millisecond * 100 * time.Duration(i+1))
		}
	}
	return err
}

func (db *SQLiteDatabase) executeTransaction(ctx context.Context, fn TransactionFunc) error {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	sqliteTx := &SQLiteTransaction{tx: tx}
	err = fn(sqliteTx)
	return err
}

func (db *SQLiteDatabase) GetConnection() interface{} {
	return db.db
}

func (db *SQLiteDatabase) GetType() DatabaseType {
	return DatabaseTypeSQLite
}

func (db *SQLiteDatabase) GetConnectionString() string {
	return db.dsn
}

func (db *SQLiteDatabase) GetMigrator() Migrator {
	return NewSQLiteMigrator(db)
}

// SQLiteTransaction implements Transaction interface
type SQLiteTransaction struct {
	tx *sql.Tx
}

func (t *SQLiteTransaction) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	result, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SQLiteResult{result: result}, nil
}

func (t *SQLiteTransaction) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SQLiteRows{rows: rows}, nil
}

func (t *SQLiteTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	row := t.tx.QueryRowContext(ctx, query, args...)
	return &SQLiteRow{row: row}
}

func (t *SQLiteTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit()
}

func (t *SQLiteTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback()
}

// SQLite-specific implementations of Result, Rows, Row
type SQLiteResult struct {
	result sql.Result
}

func (r *SQLiteResult) RowsAffected() (int64, error) {
	return r.result.RowsAffected()
}

func (r *SQLiteResult) LastInsertId() (int64, error) {
	return r.result.LastInsertId()
}

type SQLiteRows struct {
	rows *sql.Rows
}

func (r *SQLiteRows) Next() bool {
	return r.rows.Next()
}

func (r *SQLiteRows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

func (r *SQLiteRows) Close() error {
	return r.rows.Close()
}

func (r *SQLiteRows) Err() error {
	return r.rows.Err()
}

type SQLiteRow struct {
	row *sql.Row
}

func (r *SQLiteRow) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}

// SQLite migrator placeholder
type SQLiteMigrator struct {
	db *SQLiteDatabase
}

func NewSQLiteMigrator(db *SQLiteDatabase) Migrator {
	return &SQLiteMigrator{db: db}
}

func (m *SQLiteMigrator) RunMigrations(ctx context.Context, migrationsPath string) error {
	// TODO: Implement migration logic
	return fmt.Errorf("SQLite migrations not implemented yet")
}

func (m *SQLiteMigrator) GetVersion(ctx context.Context) (int, error) {
	// TODO: Implement version tracking
	return 0, fmt.Errorf("SQLite version tracking not implemented yet")
}

// Helper function to check if SQLite error is retryable
func isSQLiteRetryableError(err error) bool {
	// Check for SQLITE_BUSY, SQLITE_LOCKED, etc.
	return false // Simplified for now
}
