# Multi-Database Provider Pattern

## Overview
Document outlining the design for a multi-database provider pattern in VulnSense, enabling support for multiple database types (PostgreSQL, MySQL, SQLite, etc.) through a unified interface.

## Architecture Design

### Core Interfaces

#### 1. Database Interface
```go
// internal/database/interface.go
package database

import (
    "context"
    "time"
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
```

#### 2. Database Provider Interface
```go
// internal/database/provider.go
package database

import (
    "fmt"
    "vulnsense/internal/config"
)

// DatabaseProvider creates and configures database instances
type DatabaseProvider interface {
    CreateDatabase(config config.DatabaseConfig) (Database, error)
    GetType() DatabaseType
    ValidateConfig(config config.DatabaseConfig) error
    GetDefaultConfig() config.DatabaseConfig
}

// DatabaseType represents supported database types
type DatabaseType string

const (
    DatabaseTypePostgreSQL DatabaseType = "postgresql"
    DatabaseTypeMySQL      DatabaseType = "mysql"
    DatabaseTypeSQLite     DatabaseType = "sqlite"
    DatabaseTypeMongoDB    DatabaseType = "mongodb"
)

// DatabaseFactory manages database providers
type DatabaseFactory struct {
    providers map[DatabaseType]DatabaseProvider
}

func NewDatabaseFactory() *DatabaseFactory {
    factory := &DatabaseFactory{
        providers: make(map[DatabaseType]DatabaseProvider),
    }
    
    // Register default providers
    factory.RegisterProvider(NewPostgreSQLProvider())
    factory.RegisterProvider(NewMySQLProvider())
    factory.RegisterProvider(NewSQLiteProvider())
    
    return factory
}

func (f *DatabaseFactory) RegisterProvider(provider DatabaseProvider) {
    f.providers[provider.GetType()] = provider
}

func (f *DatabaseFactory) CreateDatabase(config config.DatabaseConfig) (Database, error) {
    provider, exists := f.providers[config.Type]
    if !exists {
        return nil, fmt.Errorf("unsupported database type: %s", config.Type)
    }
    
    if err := provider.ValidateConfig(config); err != nil {
        return nil, fmt.Errorf("invalid configuration for %s: %w", config.Type, err)
    }
    
    return provider.CreateDatabase(config)
}
```

### Implementation Examples

#### PostgreSQL Provider
```go
// internal/database/postgresql_provider.go
package database

import (
    "context"
    "fmt"
    "time"
    
    "github.com/jackc/pgx/v5/pgxpool"
    "vulnsense/internal/config"
)

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
    poolConfig.MaxConns = int32(config.MaxConnections)
    poolConfig.MinConns = int32(config.MinConnections)
    poolConfig.MaxConnLifetime = config.MaxLifetime
    poolConfig.MaxConnIdleTime = config.MaxIdleTime
    
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
    return nil
}

func (p *PostgreSQLProvider) GetDefaultConfig() config.DatabaseConfig {
    return config.DatabaseConfig{
        Type:           DatabaseTypePostgreSQL,
        Host:           "localhost",
        Port:           5432,
        SSLMode:        "require",
        MaxConnections: 10,
        MinConnections: 2,
        MaxLifetime:    time.Hour,
        MaxIdleTime:    time.Minute * 30,
    }
}

// PostgreSQLDatabase implements Database interface
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

func (db *PostgreSQLDatabase) GetConnection() interface{} {
    return db.pool
}

func (db *PostgreSQLDatabase) GetType() DatabaseType {
    return DatabaseTypePostgreSQL
}
```

#### MySQL Provider
```go
// internal/database/mysql_provider.go
package database

import (
    "database/sql"
    "fmt"
    "time"
    
    _ "github.com/go-sql-driver/mysql"
    "vulnsense/internal/config"
)

type MySQLProvider struct{}

func NewMySQLProvider() DatabaseProvider {
    return &MySQLProvider{}
}

func (p *MySQLProvider) CreateDatabase(config config.DatabaseConfig) (Database, error) {
    dsn := fmt.Sprintf(
        "%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
        config.User, config.Password, config.Host, config.Port, config.Name,
    )
    
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open MySQL connection: %w", err)
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(config.MaxConnections)
    db.SetMaxIdleConns(config.MinConnections)
    db.SetConnMaxLifetime(config.MaxLifetime)
    db.SetConnMaxIdleTime(config.MaxIdleTime)
    
    return &MySQLDatabase{
        db:     db,
        config: config,
        dsn:    dsn,
    }, nil
}
```

#### SQLite Provider
```go
// internal/database/sqlite_provider.go
package database

import (
    "database/sql"
    "fmt"
    
    _ "github.com/mattn/go-sqlite3"
    "vulnsense/internal/config"
)

type SQLiteProvider struct{}

func NewSQLiteProvider() DatabaseProvider {
    return &SQLiteProvider{}
}

func (p *SQLiteProvider) CreateDatabase(config config.DatabaseConfig) (Database, error) {
    dsn := config.Name // For SQLite, Name is the file path
    if dsn == "" {
        dsn = ":memory:" // In-memory database
    }
    
    db, err := sql.Open("sqlite3", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open SQLite connection: %w", err)
    }
    
    // SQLite-specific settings
    db.SetMaxOpenConns(1) // SQLite doesn't support concurrent writes
    
    return &SQLiteDatabase{
        db:     db,
        config: config,
        dsn:    dsn,
    }, nil
}
```

### Database Manager

```go
// internal/database/manager.go
package database

import (
    "context"
    "fmt"
    "sync"
    
    "vulnsense/internal/config"
)

// DatabaseManager manages multiple database connections
type DatabaseManager struct {
    factory   *DatabaseFactory
    databases map[string]Database
    mutex     sync.RWMutex
}

func NewDatabaseManager() *DatabaseManager {
    return &DatabaseManager{
        factory:   NewDatabaseFactory(),
        databases: make(map[string]Database),
    }
}

// InitializePrimary initializes the primary database
func (dm *DatabaseManager) InitializePrimary(config config.DatabaseConfig) (Database, error) {
    db, err := dm.factory.CreateDatabase(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create primary database: %w", err)
    }
    
    if err := db.Connect(context.Background()); err != nil {
        return nil, fmt.Errorf("failed to connect to primary database: %w", err)
    }
    
    dm.mutex.Lock()
    dm.databases["primary"] = db
    dm.mutex.Unlock()
    
    return db, nil
}

// Initialize initializes multiple databases
func (dm *DatabaseManager) Initialize(configs map[string]config.DatabaseConfig) error {
    dm.mutex.Lock()
    defer dm.mutex.Unlock()
    
    for name, config := range configs {
        db, err := dm.factory.CreateDatabase(config)
        if err != nil {
            return fmt.Errorf("failed to create database %s: %w", name, err)
        }
        
        if err := db.Connect(context.Background()); err != nil {
            return fmt.Errorf("failed to connect to database %s: %w", name, err)
        }
        
        dm.databases[name] = db
    }
    
    return nil
}

// GetDatabase returns a database by name
func (dm *DatabaseManager) GetDatabase(name string) (Database, error) {
    dm.mutex.RLock()
    defer dm.mutex.RUnlock()
    
    db, exists := dm.databases[name]
    if !exists {
        return nil, fmt.Errorf("database %s not found", name)
    }
    
    return db, nil
}

// GetPrimary returns the primary database
func (dm *DatabaseManager) GetPrimary() (Database, error) {
    return dm.GetDatabase("primary")
}

// HealthCheck checks health of all databases
func (dm *DatabaseManager) HealthCheck(ctx context.Context) map[string]*HealthStatus {
    dm.mutex.RLock()
    defer dm.mutex.RUnlock()
    
    health := make(map[string]*HealthStatus)
    for name, db := range dm.databases {
        health[name] = db.HealthCheck(ctx)
    }
    
    return health
}
```

### Configuration

```go
// internal/config/database.go
package config

import (
    "time"
    "vulnsense/internal/database"
)

// DatabaseConfig holds configuration for any database type
type DatabaseConfig struct {
    // Database type
    Type database.DatabaseType `mapstructure:"type" yaml:"type"`
    
    // Connection details
    Host     string `mapstructure:"host" yaml:"host"`
    Port     int    `mapstructure:"port" yaml:"port"`
    User     string `mapstructure:"user" yaml:"user"`
    Password string `mapstructure:"password" yaml:"password"`
    Name     string `mapstructure:"name" yaml:"name"`
    
    // PostgreSQL specific
    SSLMode string `mapstructure:"sslmode" yaml:"sslmode"`
    
    // Connection pool settings
    MaxConnections int           `mapstructure:"max_connections" yaml:"max_connections"`
    MinConnections int           `mapstructure:"min_connections" yaml:"min_connections"`
    MaxLifetime    time.Duration `mapstructure:"max_lifetime" yaml:"max_lifetime"`
    MaxIdleTime    time.Duration `mapstructure:"max_idle_time" yaml:"max_idle_time"`
    
    // Additional options
    Options map[string]interface{} `mapstructure:"options" yaml:"options"`
}

// Multi-database configuration
type Config struct {
    // Primary database
    Database DatabaseConfig `mapstructure:"database" yaml:"database"`
    
    // Optional: Multiple databases for different purposes
    Databases map[string]DatabaseConfig `mapstructure:"databases" yaml:"databases"`
}
```

### Configuration Example

```yaml
# configs/app.yaml
database:
  type: "postgresql"
  host: "localhost"
  port: 5432
  user: "vulnsense"
  password: "password"
  name: "vulnsense"
  sslmode: "require"
  max_connections: 10
  min_connections: 2
  max_lifetime: "1h"
  max_idle_time: "30m"

# Optional: Multiple databases
databases:
  primary:
    type: "postgresql"
    host: "localhost"
    port: 5432
    user: "vulnsense"
    password: "password"
    name: "vulnsense"
    
  analytics:
    type: "mysql"
    host: "analytics-db"
    port: 3306
    user: "analytics"
    password: "analytics_password"
    name: "analytics"
    
  cache:
    type: "sqlite"
    name: "/tmp/cache.db"
    
  test:
    type: "sqlite"
    name: ":memory:"
```

### Benefits
1. **Flexibility**: Support multiple database types
2. **Scalability**: Different databases for different purposes
3. **Testing**: Use SQLite for unit tests
4. **Migration Support**: Gradual migration between databases
5. **Environment-specific**: Different databases per environment

### Usage Examples

#### Single Database
```go
cfg := config.DatabaseConfig{
    Type: database.DatabaseTypePostgreSQL,
    Host: "localhost",
    Port: 5432,
    Name: "vulnsense",
}

dbManager := database.NewDatabaseManager()
db, err := dbManager.InitializePrimary(cfg)
```

#### Multi-Database Setup
```go
configs := map[string]config.DatabaseConfig{
    "primary": {
        Type: database.DatabaseTypePostgreSQL,
        Host: "primary-db",
        Name: "vulnsense",
    },
    "analytics": {
        Type: database.DatabaseTypeMySQL,
        Host: "analytics-db", 
        Name: "analytics",
    },
    "cache": {
        Type: database.DatabaseTypeSQLite,
        Name: "/tmp/cache.db",
    },
}

dbManager := database.NewDatabaseManager()
err := dbManager.Initialize(configs)

// Use different databases for different purposes
primaryDB, _ := dbManager.GetDatabase("primary")
analyticsDB, _ := dbManager.GetDatabase("analytics")
cacheDB, _ := dbManager.GetDatabase("cache")
```

#### Testing Setup
```go
// Testing with SQLite
func TestVulnerabilityRepository(t *testing.T) {
    cfg := config.DatabaseConfig{
        Type: database.DatabaseTypeSQLite,
        Name: ":memory:",
    }
    
    factory := database.NewDatabaseFactory()
    db, err := factory.CreateDatabase(cfg)
    require.NoError(t, err)
    
    repo := adapter.NewVulnerabilityRepository(db, slog.Default())
    // Test repository operations
}
```

### Repository Adaptation

```go
// internal/adapter/vulnerability_repository.go
package adapter

import (
    "context"
    "fmt"
    "log/slog"
    
    "vulnsense/internal/database"
    "vulnsense/internal/domain"
)

// VulnerabilityRepository works with any database type
type VulnerabilityRepository struct {
    db     database.Database
    logger *slog.Logger
}

func NewVulnerabilityRepository(db database.Database, logger *slog.Logger) *VulnerabilityRepository {
    return &VulnerabilityRepository{
        db:     db,
        logger: logger,
    }
}

// Save works with any database implementation
func (r *VulnerabilityRepository) Save(ctx context.Context, vuln domain.Vulnerability) error {
    query := r.getInsertQuery()
    
    return r.db.WithTransaction(ctx, func(tx database.Transaction) error {
        _, err := tx.Exec(ctx, query, 
            vuln.ID, vuln.Title, vuln.Severity, vuln.Description, vuln.PublishedDate,
        )
        if err != nil {
            r.logger.Error("Failed to save vulnerability", "error", err, "vuln_id", vuln.ID)
            return fmt.Errorf("failed to save vulnerability %s: %w", vuln.ID, err)
        }
        return nil
    })
}

// getInsertQuery returns appropriate query based on database type
func (r *VulnerabilityRepository) getInsertQuery() string {
    switch r.db.GetType() {
    case database.DatabaseTypePostgreSQL:
        return `
            INSERT INTO vulnerabilities (id, title, severity, description, published_date)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (id) DO UPDATE SET
                title = EXCLUDED.title,
                severity = EXCLUDED.severity,
                description = EXCLUDED.description,
                published_date = EXCLUDED.published_date,
                updated_at = CURRENT_TIMESTAMP
        `
    case database.DatabaseTypeMySQL:
        return `
            INSERT INTO vulnerabilities (id, title, severity, description, published_date)
            VALUES (?, ?, ?, ?, ?)
            ON DUPLICATE KEY UPDATE
                title = VALUES(title),
                severity = VALUES(severity),
                description = VALUES(description),
                published_date = VALUES(published_date),
                updated_at = CURRENT_TIMESTAMP
        `
    case database.DatabaseTypeSQLite:
        return `
            INSERT OR REPLACE INTO vulnerabilities (id, title, severity, description, published_date)
            VALUES (?, ?, ?, ?, ?)
        `
    default:
        return `
            INSERT INTO vulnerabilities (id, title, severity, description, published_date)
            VALUES (?, ?, ?, ?, ?)
        `
    }
}
```

## Migration Guide

### Phase 1: Create Interfaces
1. Create `internal/database/interface.go`
2. Create `internal/database/provider.go`
3. Create PostgreSQL provider implementation

### Phase 2: Update Application
1. Add DatabaseManager to app
2. Update repositories to use Database interface
3. Update configuration structure

### Phase 3: Add More Providers
1. Add MySQL provider
2. Add SQLite provider  
3. Add any additional database providers

### Phase 4: Multi-Database Features
1. Configure multiple databases
2. Use different databases for different purposes
3. Implement database-specific optimizations

This design provides a solid foundation for multi-database support while maintaining clean architecture and Go best practices. 