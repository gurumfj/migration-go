// Package main demonstrates basic usage patterns of the migration library.
// This example shows various ways to use the library in real applications.
package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	migration "github.com/gurumfj/migration-go"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	log.Println("🚀 Migration Library - Basic Usage Examples")

	// Example 1: Basic migration
	log.Println("=" + repeat("=", 60))
	log.Println("Example 1: Basic Migration")
	log.Println(repeat("=", 61))
	basicUsage()

	// Example 2: Check status
	log.Println("\n" + repeat("=", 61))
	log.Println("Example 2: Check Migration Status")
	log.Println(repeat("=", 61))
	checkStatus()

	// Example 3: Verbose output
	log.Println("\n" + repeat("=", 61))
	log.Println("Example 3: Verbose Migration Output")
	log.Println(repeat("=", 61))
	verboseOutput()

	// Example 4: Get current version
	log.Println("\n" + repeat("=", 61))
	log.Println("Example 4: Get Current Version")
	log.Println(repeat("=", 61))
	getCurrentVersion()

	// Example 5: Filesystem migrations
	log.Println("\n" + repeat("=", 61))
	log.Println("Example 5: Filesystem Migrations")
	log.Println(repeat("=", 61))
	filesystemMigrations()

	log.Println("\n✅ All examples completed successfully!")
}

// Example 1: Basic migration usage
func basicUsage() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create migrator with embedded migrations
	migrator := migration.NewMigratorFromFS(migrations, "migrations")

	// Run all pending migrations
	result, err := migrator.Run(db)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Printf("✓ Applied %d migrations\n", result.Applied)
	fmt.Printf("✓ Current version: %s\n", result.CurrentVersion)
}

// Example 2: Check migration status before running
func checkStatus() {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	migrator := migration.NewMigratorFromFS(migrations, "migrations")

	// Check status first
	status, err := migrator.Status(db)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Current version: %s\n", status.CurrentVersion)
	fmt.Printf("Is up to date: %v\n", status.IsUpToDate)
	fmt.Printf("Pending migrations: %d\n", len(status.PendingMigrations))
	fmt.Printf("Applied migrations: %d\n", len(status.AppliedMigrations))

	// Show pending migrations
	if len(status.PendingMigrations) > 0 {
		fmt.Println("\nPending migrations:")
		for _, m := range status.PendingMigrations {
			fmt.Printf("  - %s: %s\n", m.ID, m.Description)
		}
	}

	// Run migrations
	result, _ := migrator.Run(db)
	fmt.Printf("\n✓ Applied %d migrations\n", result.Applied)
}

// Example 3: Verbose migration output
func verboseOutput() {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	migrator := migration.NewMigratorFromFS(migrations, "migrations")

	// Run with verbose logging
	result, err := migrator.Run(db, migration.WithVerbose())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n✓ Migration completed: %d applied\n", result.Applied)
}

// Example 4: Get current version after migration
func getCurrentVersion() {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	migrator := migration.NewMigratorFromFS(migrations, "migrations")

	// Run migrations first
	migrator.Run(db)

	// Get current version
	version, err := migration.GetCurrentVersion(db)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ Current migration ID: %s\n", version)

	// Query migration history
	rows, _ := db.Query("SELECT id, description, applied_at FROM schema_migrations ORDER BY id")
	defer rows.Close()

	fmt.Println("\nMigration history:")
	for rows.Next() {
		var id, desc, appliedAt string
		rows.Scan(&id, &desc, &appliedAt)
		fmt.Printf("  - [%s] %s (applied at: %s)\n", id, desc, appliedAt)
	}
}

// Example 5: Using filesystem migrations (directory-based)
func filesystemMigrations() {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	// Create a migrator that reads from filesystem
	migrator := migration.NewMigratorFromDir("./migrations")

	// Run migrations
	result, err := migrator.Run(db)
	if err != nil {
		fmt.Printf("✗ Cannot read from filesystem: %v\n", err)
		fmt.Println("  (This is expected when migrations directory doesn't exist)")
		fmt.Println("  In real projects, use this for development environments")
		return
	}

	fmt.Printf("✓ Applied %d migrations from filesystem\n", result.Applied)
}

// Helper function to repeat strings
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
