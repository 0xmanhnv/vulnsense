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
	factory.RegisterProvider(NewSQLiteProvider())
	// factory.RegisterProvider(NewMySQLProvider()) // Add when implemented

	return factory
}

func (f *DatabaseFactory) RegisterProvider(provider DatabaseProvider) {
	f.providers[provider.GetType()] = provider
}

func (f *DatabaseFactory) CreateDatabase(config config.DatabaseConfig) (Database, error) {
	dbType := DatabaseType(config.Type)
	provider, exists := f.providers[dbType]
	if !exists {
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	if err := provider.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration for %s: %w", config.Type, err)
	}

	return provider.CreateDatabase(config)
}

func (f *DatabaseFactory) GetSupportedTypes() []DatabaseType {
	types := make([]DatabaseType, 0, len(f.providers))
	for dbType := range f.providers {
		types = append(types, dbType)
	}
	return types
}
