package database

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"vulnsense/internal/config"
)

// DatabaseManager manages multiple database connections with both raw SQL and ORM support
type DatabaseManager struct {
	factory      *DatabaseFactory
	databases    map[string]Database
	ormDatabases map[string]ORMDatabase
	mutex        sync.RWMutex
	ormMutex     sync.RWMutex
	logger       *slog.Logger
}

func NewDatabaseManager(logger *slog.Logger) *DatabaseManager {
	return &DatabaseManager{
		factory:      NewDatabaseFactory(),
		databases:    make(map[string]Database),
		ormDatabases: make(map[string]ORMDatabase),
		logger:       logger,
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

// InitializeORMDatabase creates both raw and ORM interfaces
func (dm *DatabaseManager) InitializeORMDatabase(name string, config config.DatabaseConfig) (ORMDatabase, error) {
	// First create raw database
	rawDB, err := dm.factory.CreateDatabase(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create raw database: %w", err)
	}

	if err := rawDB.Connect(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to raw database: %w", err)
	}

	// Then wrap with ORM
	ormDB, err := NewORMAdapter(rawDB, dm.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create ORM adapter: %w", err)
	}

	dm.mutex.Lock()
	dm.databases[name] = rawDB
	dm.mutex.Unlock()

	dm.ormMutex.Lock()
	dm.ormDatabases[name] = ormDB
	dm.ormMutex.Unlock()

	if dm.logger != nil {
		dm.logger.Info("ORM database initialized successfully",
			"name", name,
			"type", rawDB.GetType(),
		)
	}

	return ormDB, nil
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

// GetORMDatabase returns an ORM database by name
func (dm *DatabaseManager) GetORMDatabase(name string) (ORMDatabase, error) {
	dm.ormMutex.RLock()
	defer dm.ormMutex.RUnlock()

	db, exists := dm.ormDatabases[name]
	if !exists {
		return nil, fmt.Errorf("ORM database %s not found", name)
	}

	return db, nil
}

// GetPrimary returns the primary database
func (dm *DatabaseManager) GetPrimary() (Database, error) {
	return dm.GetDatabase("primary")
}

// GetPrimaryORM returns the primary ORM database
func (dm *DatabaseManager) GetPrimaryORM() (ORMDatabase, error) {
	return dm.GetORMDatabase("primary")
}

// ListDatabases returns the names of all registered databases
func (dm *DatabaseManager) ListDatabases() []string {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	names := make([]string, 0, len(dm.databases))
	for name := range dm.databases {
		names = append(names, name)
	}
	return names
}

// ListORMDatabases returns the names of all registered ORM databases
func (dm *DatabaseManager) ListORMDatabases() []string {
	dm.ormMutex.RLock()
	defer dm.ormMutex.RUnlock()

	names := make([]string, 0, len(dm.ormDatabases))
	for name := range dm.ormDatabases {
		names = append(names, name)
	}
	return names
}

// Close closes all database connections
func (dm *DatabaseManager) Close() error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	var errors []error
	for name, db := range dm.databases {
		if err := db.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close database %s: %w", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing databases: %v", errors)
	}

	return nil
}

// CloseORMDatabase closes a specific ORM database
func (dm *DatabaseManager) CloseORMDatabase(name string) error {
	dm.ormMutex.Lock()
	defer dm.ormMutex.Unlock()

	ormDB, exists := dm.ormDatabases[name]
	if !exists {
		return fmt.Errorf("ORM database %s not found", name)
	}

	// Close the underlying raw database
	if err := ormDB.Close(); err != nil {
		return fmt.Errorf("failed to close ORM database %s: %w", name, err)
	}

	delete(dm.ormDatabases, name)

	if dm.logger != nil {
		dm.logger.Info("ORM database closed successfully", "name", name)
	}

	return nil
}

// CloseAllORMDatabases closes all ORM databases
func (dm *DatabaseManager) CloseAllORMDatabases() error {
	dm.ormMutex.Lock()
	defer dm.ormMutex.Unlock()

	var errs []error
	for name, ormDB := range dm.ormDatabases {
		if err := ormDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close ORM database %s: %w", name, err))
		}
	}

	dm.ormDatabases = make(map[string]ORMDatabase)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing ORM databases: %v", errs)
	}

	if dm.logger != nil {
		dm.logger.Info("All ORM databases closed successfully")
	}

	return nil
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

// Stats returns statistics for all databases
func (dm *DatabaseManager) Stats() map[string]DatabaseStats {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	stats := make(map[string]DatabaseStats)
	for name, db := range dm.databases {
		stats[name] = db.Stats()
	}

	return stats
}

// GetSupportedTypes returns all supported database types
func (dm *DatabaseManager) GetSupportedTypes() []DatabaseType {
	return dm.factory.GetSupportedTypes()
}

// RegisterProvider registers a new database provider
func (dm *DatabaseManager) RegisterProvider(provider DatabaseProvider) {
	dm.factory.RegisterProvider(provider)
}

// GetOverallHealthStatus returns an overall health status for all databases
func (dm *DatabaseManager) GetOverallHealthStatus(ctx context.Context) *OverallHealthStatus {
	health := dm.HealthCheck(ctx)

	overall := &OverallHealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Databases: health,
	}

	// Check if any database is unhealthy
	for _, status := range health {
		if status.Status != "healthy" {
			overall.Status = "unhealthy"
			overall.Message = "One or more databases are unhealthy"
			break
		}
	}

	return overall
}

// OverallHealthStatus represents the health status of all databases
type OverallHealthStatus struct {
	Status    string                   `json:"status"`
	Message   string                   `json:"message,omitempty"`
	Timestamp time.Time                `json:"timestamp"`
	Databases map[string]*HealthStatus `json:"databases"`
}
