# ORM Integration Strategy

## Overview
This document outlines strategies for integrating ORMs (Object-Relational Mapping) with VulnSense's multi-database provider pattern while maintaining flexibility and performance.

## Popular Go ORMs Comparison

### 1. GORM - Most Popular Choice
```go
import (
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
    "gorm.io/driver/mysql" 
    "gorm.io/driver/sqlite"
)
```

**Pros:**
- Mature and widely adopted
- Multi-database support (PostgreSQL, MySQL, SQLite, SQL Server)
- Auto-migrations
- Rich associations and hooks
- Plugin ecosystem

**Cons:**
- Performance overhead
- Magic behavior can be unpredictable
- Learning curve for complex queries

### 2. Ent by Facebook - Type-Safe Code Generation
```go
import "entgo.io/ent"
```

**Pros:**
- Type-safe and compile-time checked
- Code generation from schema
- Graph-based querying
- Excellent documentation

**Cons:**
- Steeper learning curve
- Less mature ecosystem
- Opinionated structure

### 3. XORM - Lightweight Alternative
```go
import "xorm.io/xorm"
```

**Pros:**
- Lightweight and fast
- Simple API
- Good performance

**Cons:**
- Less feature-rich
- Smaller community

### 4. SQLBoiler - Performance-First
```go
import "github.com/volatiletech/sqlboiler/v4"
```

**Pros:**
- Excellent performance
- Code generation from existing schema
- Type-safe

**Cons:**
- Complex setup
- Less flexible for rapid development

## Integration Architecture

### Hybrid Approach - Best of Both Worlds

We can design an architecture that supports both ORM and raw SQL:

```go
// internal/database/orm_adapter.go
package database

import (
    "context"
    "gorm.io/gorm"
)

// ORMDatabase extends Database interface with ORM capabilities
type ORMDatabase interface {
    Database
    GetORM() *gorm.DB
    WithORMTransaction(ctx context.Context, fn ORMTransactionFunc) error
}

type ORMTransactionFunc func(tx *gorm.DB) error

// ORMAdapter wraps our Database with GORM
type ORMAdapter struct {
    Database
    orm *gorm.DB
}

func NewORMAdapter(db Database) (ORMDatabase, error) {
    var dialector gorm.Dialector
    
    switch db.GetType() {
    case DatabaseTypePostgreSQL:
        dialector = postgres.New(postgres.Config{
            Conn: db.GetConnection().(*pgxpool.Pool),
        })
    case DatabaseTypeMySQL:
        dialector = mysql.New(mysql.Config{
            Conn: db.GetConnection().(*sql.DB),
        })
    case DatabaseTypeSQLite:
        dialector = sqlite.Open(db.GetConnectionString())
    default:
        return nil, fmt.Errorf("unsupported database type for ORM: %s", db.GetType())
    }
    
    orm, err := gorm.Open(dialector, &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to initialize ORM: %w", err)
    }
    
    return &ORMAdapter{
        Database: db,
        orm:      orm,
    }, nil
}

func (oa *ORMAdapter) GetORM() *gorm.DB {
    return oa.orm
}

func (oa *ORMAdapter) WithORMTransaction(ctx context.Context, fn ORMTransactionFunc) error {
    return oa.orm.WithContext(ctx).Transaction(fn)
}
```

### Enhanced Database Manager with ORM Support

```go
// internal/database/orm_manager.go
package database

type EnhancedDatabaseManager struct {
    *DatabaseManager
    ormDatabases map[string]ORMDatabase
    ormMutex     sync.RWMutex
}

func NewEnhancedDatabaseManager() *EnhancedDatabaseManager {
    return &EnhancedDatabaseManager{
        DatabaseManager: NewDatabaseManager(),
        ormDatabases:    make(map[string]ORMDatabase),
    }
}

// InitializeORMDatabase creates both raw and ORM interfaces
func (edm *EnhancedDatabaseManager) InitializeORMDatabase(name string, config config.DatabaseConfig) (ORMDatabase, error) {
    // First create raw database
    rawDB, err := edm.factory.CreateDatabase(config)
    if err != nil {
        return nil, err
    }
    
    if err := rawDB.Connect(context.Background()); err != nil {
        return nil, err
    }
    
    // Then wrap with ORM
    ormDB, err := NewORMAdapter(rawDB)
    if err != nil {
        return nil, err
    }
    
    edm.mutex.Lock()
    edm.databases[name] = rawDB
    edm.mutex.Unlock()
    
    edm.ormMutex.Lock()
    edm.ormDatabases[name] = ormDB
    edm.ormMutex.Unlock()
    
    return ormDB, nil
}

func (edm *EnhancedDatabaseManager) GetORMDatabase(name string) (ORMDatabase, error) {
    edm.ormMutex.RLock()
    defer edm.ormMutex.RUnlock()
    
    db, exists := edm.ormDatabases[name]
    if !exists {
        return nil, fmt.Errorf("ORM database %s not found", name)
    }
    
    return db, nil
}
```

## Domain Models with GORM

### Vulnerability Model
```go
// internal/domain/models.go
package domain

import (
    "time"
    "gorm.io/gorm"
)

type Vulnerability struct {
    ID          string    `gorm:"primaryKey;size:255" json:"id"`
    Title       string    `gorm:"size:500;not null" json:"title"`
    Severity    string    `gorm:"size:20;not null;index" json:"severity"`
    Description string    `gorm:"type:text" json:"description"`
    CVSSScore   *float64  `gorm:"type:decimal(3,1)" json:"cvss_score"`
    
    // Metadata
    Source      string    `gorm:"size:100;not null;index" json:"source"`
    SourceURL   string    `gorm:"size:1000" json:"source_url"`
    PublishedAt time.Time `gorm:"not null;index" json:"published_at"`
    
    // Relationships
    Tags        []Tag     `gorm:"many2many:vulnerability_tags" json:"tags"`
    Assets      []Asset   `gorm:"many2many:vulnerability_assets" json:"assets"`
    
    // GORM fields
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Tag struct {
    ID           uint   `gorm:"primaryKey" json:"id"`
    Name         string `gorm:"size:100;uniqueIndex;not null" json:"name"`
    Category     string `gorm:"size:50" json:"category"`
    
    // Relationships
    Vulnerabilities []Vulnerability `gorm:"many2many:vulnerability_tags" json:"-"`
    
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type Asset struct {
    ID       uint   `gorm:"primaryKey" json:"id"`
    Name     string `gorm:"size:255;not null" json:"name"`
    Type     string `gorm:"size:50;not null" json:"type"`
    Version  string `gorm:"size:100" json:"version"`
    Vendor   string `gorm:"size:100" json:"vendor"`
    
    // Relationships  
    Vulnerabilities []Vulnerability `gorm:"many2many:vulnerability_assets" json:"-"`
    
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

## Repository Pattern with ORM Support

### Enhanced Repository Interface
```go
// internal/adapter/repository_interface.go
package adapter

import (
    "context"
    "vulnsense/internal/domain"
)

// VulnerabilityRepository supports both raw SQL and ORM
type VulnerabilityRepository interface {
    // Raw SQL methods
    Save(ctx context.Context, vuln domain.Vulnerability) error
    FindByID(ctx context.Context, id string) (*domain.Vulnerability, error)
    
    // ORM methods
    CreateWithORM(ctx context.Context, vuln *domain.Vulnerability) error
    FindWithTags(ctx context.Context, id string) (*domain.Vulnerability, error)
    SearchByKeywords(ctx context.Context, keywords []string) ([]domain.Vulnerability, error)
    GetStats(ctx context.Context) (*VulnerabilityStats, error)
}

type VulnerabilityStats struct {
    Total        int64            `json:"total"`
    BySeverity   map[string]int64 `json:"by_severity"`
    BySource     map[string]int64 `json:"by_source"`
    RecentCount  int64            `json:"recent_count"`
}
```

### Hybrid Repository Implementation
```go
// internal/adapter/vulnerability_repository_enhanced.go
package adapter

import (
    "context"
    "fmt"
    "log/slog"
    "time"
    
    "vulnsense/internal/database"
    "vulnsense/internal/domain"
    "gorm.io/gorm"
)

type EnhancedVulnerabilityRepository struct {
    rawDB  database.Database
    ormDB  database.ORMDatabase
    logger *slog.Logger
}

func NewEnhancedVulnerabilityRepository(ormDB database.ORMDatabase, logger *slog.Logger) VulnerabilityRepository {
    return &EnhancedVulnerabilityRepository{
        rawDB:  ormDB,
        ormDB:  ormDB,
        logger: logger,
    }
}

// Raw SQL method - for performance-critical operations
func (r *EnhancedVulnerabilityRepository) Save(ctx context.Context, vuln domain.Vulnerability) error {
    query := r.getUpsertQuery()
    
    return r.rawDB.WithTransaction(ctx, func(tx database.Transaction) error {
        _, err := tx.Exec(ctx, query, 
            vuln.ID, vuln.Title, vuln.Severity, vuln.Description, 
            vuln.CVSSScore, vuln.Source, vuln.SourceURL, vuln.PublishedAt,
        )
        if err != nil {
            r.logger.Error("Failed to save vulnerability", "error", err, "vuln_id", vuln.ID)
            return fmt.Errorf("failed to save vulnerability %s: %w", vuln.ID, err)
        }
        return nil
    })
}

// ORM method - for complex relationships and easy development
func (r *EnhancedVulnerabilityRepository) CreateWithORM(ctx context.Context, vuln *domain.Vulnerability) error {
    return r.ormDB.WithORMTransaction(ctx, func(tx *gorm.DB) error {
        if err := tx.Create(vuln).Error; err != nil {
            r.logger.Error("Failed to create vulnerability with ORM", "error", err, "vuln_id", vuln.ID)
            return fmt.Errorf("failed to create vulnerability %s: %w", vuln.ID, err)
        }
        return nil
    })
}

// ORM method - leveraging GORM's relationship loading
func (r *EnhancedVulnerabilityRepository) FindWithTags(ctx context.Context, id string) (*domain.Vulnerability, error) {
    var vuln domain.Vulnerability
    
    err := r.ormDB.GetORM().WithContext(ctx).
        Preload("Tags").
        Preload("Assets").
        First(&vuln, "id = ?", id).Error
        
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to find vulnerability with tags: %w", err)
    }
    
    return &vuln, nil
}

// ORM method - complex search with relationships
func (r *EnhancedVulnerabilityRepository) SearchByKeywords(ctx context.Context, keywords []string) ([]domain.Vulnerability, error) {
    var vulns []domain.Vulnerability
    
    query := r.ormDB.GetORM().WithContext(ctx).
        Preload("Tags").
        Where("deleted_at IS NULL")
    
    for _, keyword := range keywords {
        query = query.Where("title ILIKE ? OR description ILIKE ?", 
            "%"+keyword+"%", "%"+keyword+"%")
    }
    
    err := query.Find(&vulns).Error
    if err != nil {
        return nil, fmt.Errorf("failed to search vulnerabilities: %w", err)
    }
    
    return vulns, nil
}

// ORM method - leveraging GORM's aggregation capabilities
func (r *EnhancedVulnerabilityRepository) GetStats(ctx context.Context) (*VulnerabilityStats, error) {
    stats := &VulnerabilityStats{
        BySeverity: make(map[string]int64),
        BySource:   make(map[string]int64),
    }
    
    orm := r.ormDB.GetORM().WithContext(ctx)
    
    // Total count
    if err := orm.Model(&domain.Vulnerability{}).Count(&stats.Total).Error; err != nil {
        return nil, fmt.Errorf("failed to get total count: %w", err)
    }
    
    // Group by severity
    var severityStats []struct {
        Severity string
        Count    int64
    }
    if err := orm.Model(&domain.Vulnerability{}).
        Select("severity, count(*) as count").
        Group("severity").
        Scan(&severityStats).Error; err != nil {
        return nil, fmt.Errorf("failed to get severity stats: %w", err)
    }
    
    for _, stat := range severityStats {
        stats.BySeverity[stat.Severity] = stat.Count
    }
    
    // Group by source
    var sourceStats []struct {
        Source string
        Count  int64
    }
    if err := orm.Model(&domain.Vulnerability{}).
        Select("source, count(*) as count").
        Group("source").
        Scan(&sourceStats).Error; err != nil {
        return nil, fmt.Errorf("failed to get source stats: %w", err)
    }
    
    for _, stat := range sourceStats {
        stats.BySource[stat.Source] = stat.Count
    }
    
    // Recent count (last 7 days)
    weekAgo := time.Now().AddDate(0, 0, -7)
    if err := orm.Model(&domain.Vulnerability{}).
        Where("created_at >= ?", weekAgo).
        Count(&stats.RecentCount).Error; err != nil {
        return nil, fmt.Errorf("failed to get recent count: %w", err)
    }
    
    return stats, nil
}

func (r *EnhancedVulnerabilityRepository) FindByID(ctx context.Context, id string) (*domain.Vulnerability, error) {
    var vuln domain.Vulnerability
    
    err := r.ormDB.GetORM().WithContext(ctx).
        First(&vuln, "id = ?", id).Error
        
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to find vulnerability: %w", err)
    }
    
    return &vuln, nil
}

func (r *EnhancedVulnerabilityRepository) getUpsertQuery() string {
    switch r.rawDB.GetType() {
    case database.DatabaseTypePostgreSQL:
        return `
            INSERT INTO vulnerabilities (id, title, severity, description, cvss_score, source, source_url, published_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
            ON CONFLICT (id) DO UPDATE SET
                title = EXCLUDED.title,
                severity = EXCLUDED.severity,
                description = EXCLUDED.description,
                cvss_score = EXCLUDED.cvss_score,
                source = EXCLUDED.source,
                source_url = EXCLUDED.source_url,
                published_at = EXCLUDED.published_at,
                updated_at = CURRENT_TIMESTAMP
        `
    default:
        return `
            INSERT OR REPLACE INTO vulnerabilities (id, title, severity, description, cvss_score, source, source_url, published_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
        `
    }
}
```

## Auto-Migration Support

```go
// internal/database/migration.go
package database

import (
    "context"
    "fmt"
    "vulnsense/internal/domain"
    "gorm.io/gorm"
)

type ORMMigrator struct {
    orm *gorm.DB
}

func NewORMMigrator(ormDB ORMDatabase) *ORMMigrator {
    return &ORMMigrator{
        orm: ormDB.GetORM(),
    }
}

func (m *ORMMigrator) RunAutoMigration(ctx context.Context) error {
    // Auto-migrate all models
    err := m.orm.WithContext(ctx).AutoMigrate(
        &domain.Vulnerability{},
        &domain.Tag{},
        &domain.Asset{},
    )
    
    if err != nil {
        return fmt.Errorf("failed to run auto-migration: %w", err)
    }
    
    return nil
}

func (m *ORMMigrator) CreateIndexes(ctx context.Context) error {
    // Create custom indexes
    indexes := []string{
        "CREATE INDEX IF NOT EXISTS idx_vulnerabilities_published_severity ON vulnerabilities(published_at, severity)",
        "CREATE INDEX IF NOT EXISTS idx_vulnerabilities_source_created ON vulnerabilities(source, created_at)",
        "CREATE INDEX IF NOT EXISTS idx_tags_category_name ON tags(category, name)",
    }
    
    for _, index := range indexes {
        if err := m.orm.WithContext(ctx).Exec(index).Error; err != nil {
            return fmt.Errorf("failed to create index: %w", err)
        }
    }
    
    return nil
}
```

## Configuration for ORM

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
  
  # ORM specific settings
  orm:
    enabled: true
    log_level: "info"  # silent, error, warn, info
    auto_migrate: true
    slow_threshold: "200ms"
    colorful_logs: true
    
  # Advanced ORM settings
  advanced:
    prepared_stmt: true
    singular_table: false
    disable_foreign_key_constraint: false
```

## Benefits and Trade-offs

### Benefits of ORM Integration

**✅ Productivity:**
- Faster development with auto-generated queries
- Rich associations and eager loading
- Type-safe operations with compile-time checking

**✅ Maintainability:**
- Auto-migrations reduce schema drift
- Consistent query patterns
- Easier refactoring

**✅ Multi-Database Support:**
- GORM handles dialect differences
- Consistent API across database types
- Easy database switching

### Trade-offs to Consider

**⚠️ Performance:**
- ORM overhead for simple queries
- N+1 query problems if not careful
- Complex queries may be less optimal

**⚠️ Learning Curve:**
- Team needs to learn ORM patterns
- Magic behavior can be confusing
- Debugging ORM-generated queries

**⚠️ Control:**
- Less control over exact SQL generation
- Dependency on ORM's query optimization
- Potential for unexpected behavior

## Recommendation: Hybrid Approach

For VulnSense, I recommend a **hybrid approach**:

1. **Use ORM for:**
   - Complex relationships and joins
   - Development speed and maintainability
   - Admin panels and dashboards
   - Non-performance-critical operations

2. **Use Raw SQL for:**
   - High-performance bulk operations
   - Complex analytical queries
   - Performance-critical paths
   - Database-specific optimizations

3. **Architecture Benefits:**
   - Best of both worlds
   - Gradual migration path
   - Team can choose appropriate tool
   - Maintains flexibility

This hybrid approach allows teams to leverage ORM productivity while maintaining performance where needed. 