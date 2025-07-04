# Database Package Organization Best Practices

## Overview
This document outlines the recommended approach for organizing database-related code in the VulnSense project. The goal is to maintain clean architecture, separation of concerns, and reusability while following Go best practices.

## Recommended Structure

### Directory Organization
```
internal/
├── database/                # Database infrastructure layer
│   ├── postgres.go         # PostgreSQL connection & utilities
│   ├── migration.go        # Database migration logic
│   ├── transaction.go      # Transaction helpers
│   └── health.go          # Database health checks
├── adapter/                # Repository implementations
│   ├── postgres_repo.go   # Business logic repositories
│   └── ...
├── config/                 # Configuration management
│   └── config.go          # Database configuration
```

## Why This Approach?

### 1. **Clear Separation of Concerns**
- **Database package**: Infrastructure concerns (connections, migrations, transactions)
- **Adapter package**: Business logic implementations (repositories)
- **Config package**: Configuration management only

### 2. **Reusability**
- Database utilities can be shared across multiple repositories
- Centralized connection management
- Common transaction patterns

### 3. **Maintainability**
- Easy to locate database-related code
- Clear boundaries between layers
- Easier to test and mock

### 4. **Scalability**
- Easy to add new database utilities
- Support for multiple database types
- Centralized migration management

## Implementation

### 1. Database Connection Management

#### internal/database/postgres.go
```go
package database

import (
    "context"
    "fmt"
    "time"
    
    "github.com/jackc/pgx/v5/pgxpool"
    "vulnsense/internal/config"
)

// PostgresDB handles database connection and utilities
type PostgresDB struct {
    Pool   *pgxpool.Pool
    Config config.DBConfig
}

// NewPostgresDB creates a new PostgreSQL connection with optimized settings
func NewPostgresDB(cfg config.DBConfig) (*PostgresDB, error) {
    dsn := fmt.Sprintf(
        "postgres://%s:%s@%s:%d/%s?sslmode=%s",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode,
    )
    
    poolConfig, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to parse database config: %w", err)
    }
    
    // Configure connection pool settings
    poolConfig.MaxConns = 10
    poolConfig.MinConns = 2
    poolConfig.MaxConnLifetime = time.Hour
    poolConfig.MaxConnIdleTime = time.Minute * 30
    poolConfig.HealthCheckPeriod = time.Minute * 5
    
    pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create connection pool: %w", err)
    }
    
    // Verify connection
    if err := pool.Ping(context.Background()); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    return &PostgresDB{
        Pool:   pool,
        Config: cfg,
    }, nil
}

// Close closes the database connection pool
func (db *PostgresDB) Close() {
    if db.Pool != nil {
        db.Pool.Close()
    }
}

// Ping checks database connectivity
func (db *PostgresDB) Ping(ctx context.Context) error {
    return db.Pool.Ping(ctx)
}

// Stats returns connection pool statistics
func (db *PostgresDB) Stats() *pgxpool.Stat {
    return db.Pool.Stat()
}
```

### 2. Transaction Management

#### internal/database/transaction.go
```go
package database

import (
    "context"
    "fmt"
    
    "github.com/jackc/pgx/v5"
)

// WithTransaction executes a function within a database transaction
func (db *PostgresDB) WithTransaction(ctx context.Context, fn func(pgx.Tx) error) error {
    tx, err := db.Pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    
    defer func() {
        if p := recover(); p != nil {
            // Rollback on panic
            if rbErr := tx.Rollback(ctx); rbErr != nil {
                fmt.Printf("failed to rollback transaction: %v\n", rbErr)
            }
            panic(p)
        } else if err != nil {
            // Rollback on error
            if rbErr := tx.Rollback(ctx); rbErr != nil {
                fmt.Printf("failed to rollback transaction: %v\n", rbErr)
            }
        } else {
            // Commit on success
            if commitErr := tx.Commit(ctx); commitErr != nil {
                err = fmt.Errorf("failed to commit transaction: %w", commitErr)
            }
        }
    }()
    
    err = fn(tx)
    return err
}

// WithTransactionRetry executes a function within a transaction with retry logic
func (db *PostgresDB) WithTransactionRetry(ctx context.Context, maxRetries int, fn func(pgx.Tx) error) error {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        err := db.WithTransaction(ctx, fn)
        if err == nil {
            return nil // Success
        }
        
        lastErr = err
        
        // Check if error is retryable (e.g., serialization failure)
        if !isRetryableError(err) {
            break
        }
    }
    
    return fmt.Errorf("transaction failed after %d retries: %w", maxRetries, lastErr)
}

// isRetryableError checks if an error is worth retrying
func isRetryableError(err error) bool {
    // Add logic to determine if error is retryable
    // e.g., serialization failures, connection issues, etc.
    return false
}
```

### 3. Migration Management

#### internal/database/migration.go
```go
package database

import (
    "context"
    "fmt"
    "io/fs"
    "path/filepath"
    "sort"
    "strings"
    "time"
    
    "github.com/jackc/pgx/v5"
)

type Migrator struct {
    db *PostgresDB
}

func NewMigrator(db *PostgresDB) *Migrator {
    return &Migrator{db: db}
}

// RunMigrations executes all pending migrations
func (m *Migrator) RunMigrations(ctx context.Context, migrationDir string) error {
    // Create migrations table if not exists
    if err := m.createMigrationsTable(ctx); err != nil {
        return fmt.Errorf("failed to create migrations table: %w", err)
    }
    
    // Get list of applied migrations
    applied, err := m.getAppliedMigrations(ctx)
    if err != nil {
        return fmt.Errorf("failed to get applied migrations: %w", err)
    }
    
    // Execute pending migrations
    return m.executePendingMigrations(ctx, migrationDir, applied)
}

// createMigrationsTable creates the migrations tracking table
func (m *Migrator) createMigrationsTable(ctx context.Context) error {
    query := `
        CREATE TABLE IF NOT EXISTS migrations (
            version VARCHAR(255) PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `
    _, err := m.db.Pool.Exec(ctx, query)
    return err
}
```

### 4. Health Check Management

#### internal/database/health.go
```go
package database

import (
    "context"
    "fmt"
    "time"
)

// HealthStatus represents database health status
type HealthStatus struct {
    Status          string        `json:"status"`
    ResponseTime    time.Duration `json:"response_time"`
    OpenConnections int32         `json:"open_connections"`
    IdleConnections int32         `json:"idle_connections"`
    Message         string        `json:"message,omitempty"`
}

// HealthCheck performs a comprehensive health check
func (db *PostgresDB) HealthCheck(ctx context.Context) *HealthStatus {
    start := time.Now()
    
    // Basic connectivity check
    if err := db.Ping(ctx); err != nil {
        return &HealthStatus{
            Status:       "unhealthy",
            ResponseTime: time.Since(start),
            Message:      fmt.Sprintf("ping failed: %v", err),
        }
    }
    
    // Get pool statistics
    stats := db.Stats()
    
    // Check if we can execute a simple query
    var result int
    err := db.Pool.QueryRow(ctx, "SELECT 1").Scan(&result)
    if err != nil {
        return &HealthStatus{
            Status:       "unhealthy",
            ResponseTime: time.Since(start),
            Message:      fmt.Sprintf("query failed: %v", err),
        }
    }
    
    return &HealthStatus{
        Status:          "healthy",
        ResponseTime:    time.Since(start),
        OpenConnections: stats.TotalConns(),
        IdleConnections: stats.IdleConns(),
    }
}
```

## Usage in Repository Layer

### Repository Implementation
```go
// internal/adapter/postgres_repo.go
package adapter

import (
    "context"
    "fmt"
    "log/slog"
    
    "github.com/jackc/pgx/v5"
    "vulnsense/internal/database"
    "vulnsense/internal/domain"
)

type PostgresVulnerabilityRepository struct {
    db     *database.PostgresDB
    logger *slog.Logger
}

func NewPostgresVulnerabilityRepository(db *database.PostgresDB, logger *slog.Logger) *PostgresVulnerabilityRepository {
    return &PostgresVulnerabilityRepository{
        db:     db,
        logger: logger,
    }
}

// Save saves a single vulnerability
func (r *PostgresVulnerabilityRepository) Save(ctx context.Context, vuln domain.Vulnerability) error {
    query := `
        INSERT INTO vulnerabilities (id, title, severity, description, published_date)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (id) DO UPDATE SET
            title = EXCLUDED.title,
            severity = EXCLUDED.severity,
            description = EXCLUDED.description,
            published_date = EXCLUDED.published_date,
            updated_at = CURRENT_TIMESTAMP
    `
    
    _, err := r.db.Pool.Exec(ctx, query, 
        vuln.ID, vuln.Title, vuln.Severity, vuln.Description, vuln.PublishedDate,
    )
    
    if err != nil {
        r.logger.Error("Failed to save vulnerability", "error", err, "vuln_id", vuln.ID)
        return fmt.Errorf("failed to save vulnerability %s: %w", vuln.ID, err)
    }
    
    return nil
}

// SaveBatch saves multiple vulnerabilities in a transaction
func (r *PostgresVulnerabilityRepository) SaveBatch(ctx context.Context, vulns []domain.Vulnerability) error {
    return r.db.WithTransaction(ctx, func(tx pgx.Tx) error {
        query := `
            INSERT INTO vulnerabilities (id, title, severity, description, published_date)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (id) DO UPDATE SET
                title = EXCLUDED.title,
                severity = EXCLUDED.severity,
                description = EXCLUDED.description,
                published_date = EXCLUDED.published_date,
                updated_at = CURRENT_TIMESTAMP
        `
        
        for _, vuln := range vulns {
            _, err := tx.Exec(ctx, query, 
                vuln.ID, vuln.Title, vuln.Severity, vuln.Description, vuln.PublishedDate,
            )
            if err != nil {
                r.logger.Error("Failed to save vulnerability in batch", "error", err, "vuln_id", vuln.ID)
                return fmt.Errorf("failed to save vulnerability %s: %w", vuln.ID, err)
            }
        }
        
        return nil
    })
}
```

## Application Assembly

### Dependency Injection
```go
// internal/app/app.go
package app

import (
    "context"
    "fmt"
    "log/slog"
    "time"
    
    "vulnsense/internal/adapter"
    "vulnsense/internal/config"
    "vulnsense/internal/database"
    "vulnsense/internal/usecase"
)

type App struct {
    db         *database.PostgresDB
    fetchVulns *usecase.FetchVulnerabilitiesUseCase
    logger     *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger) (*App, error) {
    // Initialize database
    db, err := database.NewPostgresDB(cfg.DB)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize database: %w", err)
    }
    
    // Wait for database to be ready
    if err := db.WaitForConnection(context.Background(), 30*time.Second); err != nil {
        return nil, fmt.Errorf("database not ready: %w", err)
    }
    
    // Run migrations
    migrator := database.NewMigrator(db)
    if err := migrator.RunMigrations(context.Background(), "migrations"); err != nil {
        return nil, fmt.Errorf("failed to run migrations: %w", err)
    }
    
    // Initialize repositories
    vulnRepo := adapter.NewPostgresVulnerabilityRepository(db, logger)
    assetRepo := adapter.NewPostgresAssetRepository(db, logger)
    findingRepo := adapter.NewPostgresFindingRepository(db, logger)
    
    // Initialize use cases
    fetchVulns := usecase.NewFetchVulnerabilitiesUseCase(providers, vulnRepo, logger)
    
    return &App{
        db:         db,
        fetchVulns: fetchVulns,
        logger:     logger,
    }, nil
}

// Close gracefully shuts down the application
func (a *App) Close() error {
    a.logger.Info("Shutting down application...")
    
    if a.db != nil {
        a.logger.Info("Closing database connection...")
        a.db.Close()
    }
    
    return nil
}
```

## Benefits of This Approach

### 1. **Separation of Concerns**
- Database infrastructure is separate from business logic
- Clear boundaries between layers
- Easy to maintain and extend

### 2. **Reusability**
- Database utilities can be shared across repositories
- Consistent transaction handling
- Centralized connection management

### 3. **Testability**
- Easy to mock database layer for unit tests
- Consistent testing patterns
- Isolated test environments

### 4. **Maintainability**
- Clear code organization
- Easy to locate database-related code
- Consistent error handling patterns

### 5. **Scalability**
- Easy to add new database utilities
- Support for multiple database types
- Centralized configuration management

## Best Practices

### 1. **Connection Management**
- Use connection pooling with appropriate limits
- Set connection timeouts and idle timeouts
- Monitor connection pool statistics

### 2. **Error Handling**
- Always handle database errors gracefully
- Log errors with appropriate context
- Use proper error wrapping

### 3. **Transactions**
- Use transactions for multi-step operations
- Always handle rollback in defer statements
- Consider retry logic for transient failures

### 4. **Migrations**
- Version control your migrations
- Make migrations backwards compatible when possible
- Test migrations in staging environments

### 5. **Testing**
- Use separate test databases
- Clean up test data after each test
- Test error scenarios

---

This structure follows Clean Architecture principles and Go best practices, ensuring maintainable, testable, and scalable database operations. 