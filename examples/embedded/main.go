// Package main demonstrates using the migration library with embedded migrations
// in your own project.
package main

import (
	"database/sql"
	"embed"
	"log"

	migration "github.com/gurumfj/migration-go"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	log.Println("🚀 Starting application with embedded migrations...")

	// Open database connection
	db, err := sql.Open("sqlite3", "./myapp.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create migrator with embedded migrations
	migrator := migration.NewMigratorFromFS(migrations, "migrations")

	// Check current status
	log.Println("📊 Checking migration status...")
	status, err := migrator.Status(db)
	if err != nil {
		log.Fatalf("Failed to check status: %v", err)
	}

	log.Printf("Current version: %s", status.CurrentVersion)
	log.Printf("Pending migrations: %d", len(status.PendingMigrations))
	log.Printf("Applied migrations: %d", len(status.AppliedMigrations))

	// Run migrations
	if !status.IsUpToDate {
		log.Println("🔄 Running pending migrations...")
		result, err := migrator.Run(db, migration.WithVerbose())
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

		log.Printf("✅ Applied %d migrations", result.Applied)
		log.Printf("📋 Current version: %s", result.CurrentVersion)
	} else {
		log.Println("✅ Database is already up to date!")
	}

	// Your application logic starts here...
	log.Println("🎉 Application started successfully!")

	// Example: Query the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Printf("Query example: %v", err)
	} else {
		log.Printf("Users table has %d records", count)
	}
}
