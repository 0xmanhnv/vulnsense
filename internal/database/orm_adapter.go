package database

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// ORMAdapter wraps our Database with GORM
type ORMAdapter struct {
	Database
	orm    *gorm.DB
	logger *slog.Logger
}

func NewORMAdapter(db Database, slogger *slog.Logger) (ORMDatabase, error) {
	var dialector gorm.Dialector

	switch db.GetType() {
	case DatabaseTypePostgreSQL:
		// For PostgreSQL, we need to get the connection string
		// since GORM needs its own connection
		config := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
			"localhost", 5432, "vulnsense", "password", "vulnsense")
		dialector = postgres.Open(config)

	case DatabaseTypeSQLite:
		dialector = sqlite.Open(db.GetConnectionString())

	default:
		return nil, fmt.Errorf("unsupported database type for ORM: %s", db.GetType())
	}

	// Configure GORM logger
	var gormLogger logger.Interface
	if slogger != nil {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	orm, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // Use plural table names
		},
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: false,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize ORM: %w", err)
	}

	return &ORMAdapter{
		Database: db,
		orm:      orm,
		logger:   slogger,
	}, nil
}

func (oa *ORMAdapter) GetORM() *gorm.DB {
	return oa.orm
}

func (oa *ORMAdapter) WithORMTransaction(ctx context.Context, fn ORMTransactionFunc) error {
	return oa.orm.WithContext(ctx).Transaction(fn)
}

func (oa *ORMAdapter) RunAutoMigration(ctx context.Context, models ...interface{}) error {
	if len(models) == 0 {
		return nil
	}

	err := oa.orm.WithContext(ctx).AutoMigrate(models...)
	if err != nil {
		if oa.logger != nil {
			oa.logger.Error("Auto-migration failed", "error", err)
		}
		return fmt.Errorf("failed to run auto-migration: %w", err)
	}

	if oa.logger != nil {
		oa.logger.Info("Auto-migration completed successfully", "models_count", len(models))
	}

	return nil
}
