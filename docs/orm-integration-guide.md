# ORM Integration Guide for VulnSense

## Overview

This document provides a comprehensive guide for integrating ORM (Object-Relational Mapping) into VulnSense's multi-database architecture. We recommend a **hybrid approach** that leverages both ORM productivity and raw SQL performance.

## Why ORM for VulnSense?

### Current Challenges
- **Complex Relationships**: Vulnerabilities, assets, tags, and references have intricate relationships
- **Multi-Database Support**: Need consistent API across PostgreSQL, MySQL, SQLite
- **Development Speed**: Manual SQL writing slows down feature development
- **Type Safety**: Raw SQL queries lack compile-time validation

### ORM Benefits
- **Productivity**: Auto-generated queries, migrations, and relationships
- **Type Safety**: Compile-time validation and IntelliSense support
- **Multi-Database**: Single codebase works across different databases
- **Maintenance**: Consistent patterns and easier refactoring

## GORM Integration Architecture

### 1. Enhanced Database Manager

```go
// Hybrid approach supporting both raw SQL and ORM
type EnhancedDatabaseManager struct {
    *DatabaseManager              // Existing raw SQL support
    ormDatabases map[string]ORMDatabase // ORM-enabled databases
    logger       *slog.Logger
}

// Initialize database with both raw and ORM capabilities
func (edm *EnhancedDatabaseManager) InitializeORMDatabase(name string, config config.DatabaseConfig) (ORMDatabase, error)
```

### 2. ORM Database Interface

```go
type ORMDatabase interface {
    Database                                                    // Extends existing interface
    GetORM() *gorm.DB                                          // Access to GORM instance
    WithORMTransaction(ctx context.Context, fn ORMTransactionFunc) error // ORM transactions
    RunAutoMigration(ctx context.Context, models ...interface{}) error  // Auto-migration
}
```

### 3. Domain Model Enhancement

Add GORM annotations to existing domain models:

```go
// Enhanced Vulnerability model with GORM tags
type Vulnerability struct {
    ID          string    `gorm:"primaryKey;size:255" json:"id"`
    Title       string    `gorm:"size:500;not null" json:"title"`
    Severity    string    `gorm:"size:20;not null;index" json:"severity"`
    Description string    `gorm:"type:text" json:"description"`
    
    // Relationships
    Tags    []Tag   `gorm:"many2many:vulnerability_tags" json:"tags"`
    Assets  []Asset `gorm:"many2many:vulnerability_assets" json:"assets"`
    
    // GORM managed fields
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

## Implementation Strategy

### Phase 1: Foundation (Current)
- ✅ Multi-database provider pattern
- ✅ ORM adapter layer
- ✅ GORM dependency integration
- ✅ Test application

### Phase 2: Domain Model Enhancement
```go
// Add GORM annotations to existing models
type Vulnerability struct {
    // Existing fields...
    
    // Add GORM tags for better database mapping
    ID          string `gorm:"primaryKey;size:255"`
    Title       string `gorm:"size:500;not null"`
    Severity    string `gorm:"size:20;not null;index"`
    Description string `gorm:"type:text"`
    
    // Relationships (new)
    Tags        []Tag     `gorm:"many2many:vulnerability_tags"`
    Assets      []Asset   `gorm:"many2many:vulnerability_assets"`
    References  []Reference `gorm:"foreignKey:VulnerabilityID"`
    
    // GORM managed timestamps
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// New relationship models
type Tag struct {
    ID       uint   `gorm:"primaryKey"`
    Name     string `gorm:"size:100;uniqueIndex;not null"`
    Category string `gorm:"size:50"`
    
    Vulnerabilities []Vulnerability `gorm:"many2many:vulnerability_tags"`
}

type Asset struct {
    ID      uint   `gorm:"primaryKey"`
    Name    string `gorm:"size:255;not null"`
    Type    string `gorm:"size:50;not null"`
    Version string `gorm:"size:100"`
    
    Vulnerabilities []Vulnerability `gorm:"many2many:vulnerability_assets"`
}
```

### Phase 3: Repository Enhancement
```go
// Hybrid repository supporting both approaches
type VulnerabilityRepository interface {
    // Raw SQL methods (for performance)
    BulkInsert(ctx context.Context, vulns []domain.Vulnerability) error
    GetStatistics(ctx context.Context) (*VulnerabilityStats, error)
    
    // ORM methods (for productivity)
    CreateWithRelations(ctx context.Context, vuln *domain.Vulnerability) error
    FindWithTags(ctx context.Context, id string) (*domain.Vulnerability, error)
    SearchByKeywords(ctx context.Context, keywords []string) ([]domain.Vulnerability, error)
}
```

## Usage Examples

### 1. Basic CRUD Operations

```go
// Create vulnerability with ORM
func (r *VulnerabilityRepository) CreateWithORM(ctx context.Context, vuln *domain.Vulnerability) error {
    return r.ormDB.WithORMTransaction(ctx, func(tx *gorm.DB) error {
        return tx.Create(vuln).Error
    })
}

// Complex queries with relationships
func (r *VulnerabilityRepository) FindWithTags(ctx context.Context, id string) (*domain.Vulnerability, error) {
    var vuln domain.Vulnerability
    err := r.ormDB.GetORM().WithContext(ctx).
        Preload("Tags").
        Preload("Assets").
        First(&vuln, "id = ?", id).Error
    return &vuln, err
}
```

### 2. Performance-Critical Operations

```go
// Use raw SQL for bulk operations
func (r *VulnerabilityRepository) BulkInsert(ctx context.Context, vulns []domain.Vulnerability) error {
    return r.rawDB.WithTransaction(ctx, func(tx database.Transaction) error {
        query := `INSERT INTO vulnerabilities (id, title, severity, ...) VALUES`
        // Raw SQL for maximum performance
        return tx.ExecBatch(ctx, query, vulns)
    })
}
```

### 3. Advanced Queries

```go
// ORM for complex relationship queries
func (r *VulnerabilityRepository) GetCriticalVulnerabilitiesWithAssets(ctx context.Context) ([]domain.Vulnerability, error) {
    var vulns []domain.Vulnerability
    return vulns, r.ormDB.GetORM().WithContext(ctx).
        Preload("Assets").
        Preload("Tags").
        Where("severity = ?", "CRITICAL").
        Where("exploit_available = ?", true).
        Order("published_date DESC").
        Find(&vulns).Error
}
```

## Migration Strategy

### Auto-Migration
```go
// Run auto-migration for all models
func RunMigration(ormDB database.ORMDatabase) error {
    return ormDB.RunAutoMigration(context.Background(),
        &domain.Vulnerability{},
        &domain.Tag{},
        &domain.Asset{},
        &domain.Reference{},
    )
}
```

### Custom Migrations
```go
// Custom migration for complex changes
func (m *Migrator) CreateIndexes(ctx context.Context) error {
    indexes := []string{
        "CREATE INDEX IF NOT EXISTS idx_vulnerabilities_severity_date ON vulnerabilities(severity, published_date)",
        "CREATE INDEX IF NOT EXISTS idx_vulnerabilities_source_created ON vulnerabilities(source, created_at)",
    }
    
    for _, index := range indexes {
        if err := m.orm.WithContext(ctx).Exec(index).Error; err != nil {
            return err
        }
    }
    return nil
}
```

## Configuration

### Database Configuration
```yaml
# configs/app.yaml
database:
  type: "postgresql"
  host: "localhost"
  port: 5432
  user: "vulnsense"
  password: "password"
  name: "vulnsense"
  
  # ORM settings
  orm:
    enabled: true
    log_level: "info"
    auto_migrate: true
    slow_threshold: "200ms"
    
  # Advanced settings
  advanced:
    prepared_stmt: true
    singular_table: false
    disable_foreign_key_constraint: false
```

### Application Configuration
```go
// Initialize ORM-enabled database
func InitializeApplication() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    manager := database.NewEnhancedDatabaseManager(logger)
    
    // Load configuration
    cfg, err := config.NewConfig()
    if err != nil {
        log.Fatal(err)
    }
    
    // Initialize ORM database
    ormDB, err := manager.InitializeORMDatabase("main", cfg.Database)
    if err != nil {
        log.Fatal(err)
    }
    
    // Run migrations
    if err := ormDB.RunAutoMigration(context.Background(), domain.GetAllModels()...); err != nil {
        log.Fatal(err)
    }
}
```

## Performance Considerations

### When to Use ORM
- **Development Phase**: Rapid prototyping and feature development
- **Complex Relationships**: Multi-table joins and associations
- **Admin Interfaces**: CRUD operations for management panels
- **Reporting**: Complex queries with aggregations

### When to Use Raw SQL
- **High Performance**: Bulk operations and data processing
- **Complex Analytics**: Advanced SQL features and optimizations
- **RSS Feed Processing**: Bulk inserts from external sources
- **Real-time Operations**: Low-latency requirements

### Hybrid Example
```go
// Use ORM for complex relationships
func (s *VulnerabilityService) GetVulnerabilityDetails(ctx context.Context, id string) (*VulnerabilityDetails, error) {
    var vuln domain.Vulnerability
    err := s.ormRepo.GetORM().WithContext(ctx).
        Preload("Tags").
        Preload("Assets").
        Preload("References").
        First(&vuln, "id = ?", id).Error
    
    if err != nil {
        return nil, err
    }
    
    // Use raw SQL for performance-critical statistics
    stats, err := s.rawRepo.GetVulnerabilityStatistics(ctx, id)
    if err != nil {
        return nil, err
    }
    
    return &VulnerabilityDetails{
        Vulnerability: vuln,
        Statistics:    stats,
    }, nil
}
```

## Testing Strategy

### Unit Tests
```go
func TestVulnerabilityRepository_CreateWithORM(t *testing.T) {
    // Setup in-memory SQLite for testing
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    
    // Auto-migrate
    err = db.AutoMigrate(&domain.Vulnerability{}, &domain.Tag{})
    require.NoError(t, err)
    
    // Test creation
    vuln := &domain.Vulnerability{
        ID:       "CVE-2024-TEST",
        Title:    "Test Vulnerability",
        Severity: domain.SeverityHigh,
    }
    
    err = repo.CreateWithORM(context.Background(), vuln)
    assert.NoError(t, err)
    assert.NotEmpty(t, vuln.CreatedAt)
}
```

### Integration Tests
```go
func TestORMIntegration(t *testing.T) {
    // Test with real database
    manager := database.NewEnhancedDatabaseManager(logger)
    
    ormDB, err := manager.InitializeORMDatabase("test", config.DatabaseConfig{
        Type: "sqlite",
        Name: "test.db",
    })
    require.NoError(t, err)
    
    // Test operations
    // ... integration test scenarios
}
```

## Recommended Implementation Plan

### Week 1-2: Foundation
1. ✅ Complete ORM adapter implementation
2. ✅ Fix GORM logger configuration
3. ✅ Create test application
4. ✅ Add comprehensive documentation

### Week 3-4: Domain Enhancement
1. Add GORM tags to existing domain models
2. Create relationship models (Tag, Asset, Reference)
3. Implement auto-migration system
4. Create migration utilities

### Week 5-6: Repository Layer
1. Enhance repository interfaces
2. Implement hybrid repository pattern
3. Add ORM-specific methods
4. Performance optimization

### Week 7-8: Integration & Testing
1. Update use cases to leverage ORM
2. Create comprehensive test suite
3. Performance benchmarking
4. Documentation updates

## Benefits Summary

### Development Benefits
- **40-60% faster development** for CRUD operations
- **Type safety** and compile-time validation
- **Consistent patterns** across different databases
- **Easier testing** with in-memory databases

### Operational Benefits
- **Auto-migration** reduces deployment complexity
- **Relationship management** ensures data consistency
- **Query optimization** through GORM's built-in features
- **Monitoring** through ORM query logging

### Team Benefits
- **Lower learning curve** for new team members
- **Consistent code patterns** across the codebase
- **Better IDE support** with auto-completion
- **Reduced SQL debugging** time

## Conclusion

The ORM integration provides VulnSense with:

1. **Productivity**: Faster development cycles
2. **Maintainability**: Consistent patterns and type safety
3. **Flexibility**: Choice between ORM and raw SQL
4. **Scalability**: Multi-database support with unified API

The hybrid approach ensures you get the best of both worlds: ORM productivity for development and raw SQL performance for critical operations.

**Next Steps**: 
1. Review and approve the ORM integration approach
2. Plan the domain model enhancement phase
3. Begin implementing GORM tags on existing models
4. Set up development environment with ORM testing 