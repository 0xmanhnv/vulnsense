package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"vulnsense/internal/config"
	"vulnsense/internal/database"
)

func main() {
	fmt.Println("=== Multi-Database Provider Pattern Demo ===")

	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create database manager with logger
	manager := database.NewDatabaseManager(logger)

	fmt.Printf("Supported database types: %v\n\n", manager.GetSupportedTypes())

	// Test 1: Primary PostgreSQL Database (will fail if not running)
	fmt.Println("1. Testing Primary PostgreSQL Database:")
	postgresConfig := config.DatabaseConfig{
		Type:     "postgresql",
		Host:     "localhost",
		Port:     5432,
		User:     "vulnsense",
		Password: "password",
		Name:     "vulnsense",
		SSLMode:  "require",
	}

	_, err := manager.InitializePrimary(postgresConfig)
	if err != nil {
		fmt.Printf("Failed to initialize PostgreSQL (this is expected if not running): %v\n\n", err)
	} else {
		fmt.Println("✅ PostgreSQL primary database initialized successfully\n")
	}

	// Test 2: Multiple Database Setup
	fmt.Println("2. Testing Multiple Database Setup:")
	testConfigs := map[string]config.DatabaseConfig{
		"test": {
			Type: "sqlite",
			Name: "test.db",
		},
		"analytics": {
			Type: "sqlite",
			Name: "analytics.db",
		},
		"cache": {
			Type: "sqlite",
			Name: ":memory:",
		},
	}

	if err := manager.Initialize(testConfigs); err != nil {
		log.Fatalf("Failed to initialize databases: %v", err)
	}

	databases := manager.ListDatabases()
	fmt.Printf("✅ Successfully initialized %d databases\n", len(databases))
	fmt.Printf("   Registered databases: %v\n", databases)

	// Health check
	health := manager.HealthCheck(context.Background())
	for name, status := range health {
		fmt.Printf("   Database '%s': %s (%s)\n", name, status.Status, getDatabaseType(manager, name))
	}

	// Test 3: Overall Health Check
	fmt.Println("\n3. Overall Health Check:")
	overallHealth := manager.GetOverallHealthStatus(context.Background())
	fmt.Printf("Overall Status: %s\n", overallHealth.Status)
	fmt.Printf("Timestamp: %s\n", overallHealth.Timestamp.Format("2006-01-02T15:04:05Z07:00"))

	// Test 4: Database-Specific Features
	fmt.Println("\n4. Database-Specific Features:")
	testSQLiteFeatures(manager)

	// Test 5: Provider Pattern Flexibility
	fmt.Println("\n5. Testing Provider Pattern Flexibility:")
	testProviderFlexibility(manager)

	// Clean up
	fmt.Println("\n6. Cleanup:")
	if err := manager.Close(); err != nil {
		log.Fatalf("Failed to close databases: %v", err)
	}
	fmt.Println("✅ All databases closed successfully")

	// Clean up test files
	os.Remove("test.db")
	os.Remove("analytics.db")

	fmt.Println("\n=== Demo completed ===")
}

func getDatabaseType(manager *database.DatabaseManager, name string) string {
	db, err := manager.GetDatabase(name)
	if err != nil {
		return "unknown"
	}
	return string(db.GetType())
}

func testSQLiteFeatures(manager *database.DatabaseManager) {
	fmt.Println("Testing SQLite in-memory database:")

	db, err := manager.GetDatabase("cache")
	if err != nil {
		fmt.Printf("Failed to get cache database: %v\n", err)
		return
	}

	// Test transaction
	err = db.WithTransaction(context.Background(), func(tx database.Transaction) error {
		// Create a simple table
		_, err := tx.Exec(context.Background(),
			"CREATE TABLE IF NOT EXISTS test_table (id INTEGER PRIMARY KEY, name TEXT)")
		if err != nil {
			return err
		}

		// Insert test data
		_, err = tx.Exec(context.Background(),
			"INSERT INTO test_table (name) VALUES (?)", "test_record")
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Failed SQLite transaction: %v\n", err)
	} else {
		fmt.Println("   ✅ SQLite transaction completed successfully")
	}

	// Display stats
	stats := db.Stats()
	fmt.Printf("   SQLite Stats - Open: %d, Idle: %d, Max: %d\n",
		stats.OpenConnections, stats.IdleConnections, stats.MaxConnections)
}

func testProviderFlexibility(manager *database.DatabaseManager) {
	// Test getting default configurations
	postgresProvider := database.NewPostgreSQLProvider()
	sqliteProvider := database.NewSQLiteProvider()

	postgresConfig := postgresProvider.GetDefaultConfig()
	sqliteConfig := sqliteProvider.GetDefaultConfig()

	fmt.Printf("PostgreSQL default config: Type=%s, Host=%s, Port=%d\n",
		postgresConfig.Type, postgresConfig.Host, postgresConfig.Port)
	fmt.Printf("SQLite default config: Type=%s, Name=%s\n",
		sqliteConfig.Type, sqliteConfig.Name)

	// Test config validation
	invalidConfig := config.DatabaseConfig{
		Type: "postgresql",
		// Missing required fields
	}

	err := postgresProvider.ValidateConfig(invalidConfig)
	if err != nil {
		fmt.Printf("✅ Configuration validation working: %v\n", err)
	}
}
