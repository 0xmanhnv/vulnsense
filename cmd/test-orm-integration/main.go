package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"vulnsense/internal/config"
	"vulnsense/internal/database"
	"vulnsense/internal/domain"

	"gorm.io/gorm"
)

func main() {
	fmt.Println("=== ORM Integration Test for VulnSense ===\n")

	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create unified database manager (now includes ORM support)
	manager := database.NewDatabaseManager(logger)

	// Test database configurations
	testConfigs := map[string]config.DatabaseConfig{
		"sqlite_test": {
			Type: "sqlite",
			Name: "test_orm.db",
		},
		"sqlite_analytics": {
			Type: "sqlite",
			Name: "analytics_orm.db",
		},
	}

	// Initialize databases with ORM support
	var ormDatabases []database.ORMDatabase
	for name, cfg := range testConfigs {
		fmt.Printf("Initializing ORM database: %s\n", name)

		ormDB, err := manager.InitializeORMDatabase(name, cfg)
		if err != nil {
			log.Fatalf("Failed to initialize ORM database %s: %v", name, err)
		}

		ormDatabases = append(ormDatabases, ormDB)
		fmt.Printf("✅ ORM database '%s' initialized successfully\n", name)
	}

	// Test ORM operations
	fmt.Println("\n=== Testing ORM Operations ===")

	testDB := ormDatabases[0]

	// Run auto-migration (Note: this would need actual domain models with GORM tags)
	fmt.Println("Testing auto-migration...")

	// Create a simple test struct for demonstration
	type TestVulnerability struct {
		ID        uint      `gorm:"primaryKey" json:"id"`
		Title     string    `gorm:"size:500;not null" json:"title"`
		Severity  string    `gorm:"size:20;not null;index" json:"severity"`
		Source    string    `gorm:"size:100;not null;index" json:"source"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	// Run migration
	if err := testDB.RunAutoMigration(context.Background(), &TestVulnerability{}); err != nil {
		log.Fatalf("Failed to run auto-migration: %v", err)
	}
	fmt.Println("✅ Auto-migration completed successfully")

	// Test basic CRUD operations
	fmt.Println("\n=== Testing CRUD Operations ===")

	// Create test vulnerability
	testVuln := TestVulnerability{
		Title:    "Test SQL Injection Vulnerability",
		Severity: "HIGH",
		Source:   "VulnSense-Test",
	}

	// Test ORM transaction
	err := testDB.WithORMTransaction(context.Background(), func(tx *gorm.DB) error {
		if err := tx.Create(&testVuln).Error; err != nil {
			return fmt.Errorf("failed to create test vulnerability: %w", err)
		}
		fmt.Printf("✅ Created test vulnerability with ID: %d\n", testVuln.ID)
		return nil
	})

	if err != nil {
		log.Fatalf("Failed ORM transaction: %v", err)
	}

	// Test query operations
	var count int64
	if err := testDB.GetORM().Model(&TestVulnerability{}).Count(&count).Error; err != nil {
		log.Fatalf("Failed to count vulnerabilities: %v", err)
	}
	fmt.Printf("✅ Total vulnerabilities in database: %d\n", count)

	// Test finding records
	var foundVuln TestVulnerability
	if err := testDB.GetORM().Where("severity = ?", "HIGH").First(&foundVuln).Error; err != nil {
		log.Fatalf("Failed to find vulnerability: %v", err)
	}
	fmt.Printf("✅ Found vulnerability: %s (ID: %d)\n", foundVuln.Title, foundVuln.ID)

	// Show database statistics
	fmt.Println("\n=== Database Statistics ===")
	databases := manager.ListORMDatabases()
	fmt.Printf("Active ORM databases: %v\n", databases)

	for _, name := range databases {
		db, err := manager.GetORMDatabase(name)
		if err != nil {
			continue
		}
		fmt.Printf("Database '%s': %s\n", name, db.GetType())
	}

	// Test existing domain models integration
	fmt.Println("\n=== Testing Domain Models Integration ===")

	// Create a sample vulnerability using existing domain model
	sampleVuln := domain.Vulnerability{
		ID:               "CVE-2024-TEST-001",
		Source:           "VulnSense-ORM-Test",
		Product:          domain.Product{Name: "test-app", Version: "1.0.0"},
		Title:            "Test ORM Integration Vulnerability",
		Description:      "This is a test vulnerability to demonstrate ORM integration",
		Severity:         domain.SeverityHigh,
		PublishedDate:    time.Now(),
		LastModifiedDate: time.Now(),
		References:       []string{"https://example.com/test-vuln"},
		ExploitAvailable: false,
		AffectedVersions: []string{"1.0.0"},
		Tags:             []string{"test", "orm", "integration"},
	}

	fmt.Printf("✅ Created sample domain vulnerability: %s\n", sampleVuln.ID)
	fmt.Printf("   Title: %s\n", sampleVuln.Title)
	fmt.Printf("   Severity: %s\n", sampleVuln.Severity)
	fmt.Printf("   Product: %s v%s\n", sampleVuln.Product.Name, sampleVuln.Product.Version)
	fmt.Printf("   Is Critical: %v\n", sampleVuln.IsCritical())

	// Test hybrid access (both raw SQL and ORM from same manager)
	fmt.Println("\n=== Testing Hybrid Access ===")

	// Get raw SQL database
	rawDB, err := manager.GetDatabase("sqlite_test")
	if err != nil {
		log.Fatalf("Failed to get raw database: %v", err)
	}
	fmt.Printf("✅ Raw SQL database: %s\n", rawDB.GetType())

	// Get ORM database
	ormDB, err := manager.GetORMDatabase("sqlite_test")
	if err != nil {
		log.Fatalf("Failed to get ORM database: %v", err)
	}
	fmt.Printf("✅ ORM database: %s\n", ormDB.GetType())

	fmt.Println("✅ Hybrid access working - same manager provides both raw SQL and ORM interfaces!")

	// Cleanup
	fmt.Println("\n=== Cleanup ===")
	if err := manager.CloseAllORMDatabases(); err != nil {
		log.Fatalf("Failed to close ORM databases: %v", err)
	}
	fmt.Println("✅ All ORM databases closed successfully")

	if err := manager.Close(); err != nil {
		log.Fatalf("Failed to close raw databases: %v", err)
	}
	fmt.Println("✅ All raw databases closed successfully")

	// Clean up test files
	os.Remove("test_orm.db")
	os.Remove("analytics_orm.db")
	fmt.Println("✅ Test database files cleaned up")

	fmt.Println("\n=== ORM Integration Test Completed Successfully! ===")
	fmt.Println("🎉 Unified DatabaseManager provides both raw SQL and ORM capabilities!")
}
