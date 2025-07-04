package database

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Database defines the common interface for all database implementations
type Database interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error

	// Health and monitoring
	HealthCheck(ctx context.Context) *HealthStatus
	Stats() DatabaseStats

	// Transaction management
	WithTransaction(ctx context.Context, fn TransactionFunc) error
	WithTransactionRetry(ctx context.Context, maxRetries int, fn TransactionFunc) error

	// Migration support
	GetMigrator() Migrator

	// Database-specific connection (for repositories)
	GetConnection() interface{}

	// Metadata
	GetType() DatabaseType
	GetConnectionString() string
}

// ORMDatabase extends Database interface with ORM capabilities
type ORMDatabase interface {
	Database
	GetORM() *gorm.DB
	WithORMTransaction(ctx context.Context, fn ORMTransactionFunc) error
	RunAutoMigration(ctx context.Context, models ...interface{}) error
}

type ORMTransactionFunc func(tx *gorm.DB) error

// Supporting types
type HealthStatus struct {
	Status          string        `json:"status"`
	ResponseTime    time.Duration `json:"response_time"`
	OpenConnections int32         `json:"open_connections"`
	IdleConnections int32         `json:"idle_connections"`
	Message         string        `json:"message,omitempty"`
}

type DatabaseStats struct {
	OpenConnections     int32 `json:"open_connections"`
	IdleConnections     int32 `json:"idle_connections"`
	AcquiredConnections int32 `json:"acquired_connections"`
	MaxConnections      int32 `json:"max_connections"`
}

type TransactionFunc func(tx Transaction) error

type Transaction interface {
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Database-agnostic interfaces
type Result interface {
	RowsAffected() (int64, error)
	LastInsertId() (int64, error)
}

type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	Err() error
}

type Row interface {
	Scan(dest ...interface{}) error
}

type Migrator interface {
	RunMigrations(ctx context.Context, migrationsPath string) error
	GetVersion(ctx context.Context) (int, error)
}

type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypeSQLite     DatabaseType = "sqlite"
	DatabaseTypeMongoDB    DatabaseType = "mongodb"
)
