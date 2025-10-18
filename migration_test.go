package migration

import (
	"database/sql"
	"embed"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed testdata/*.sql
var testMigrations embed.FS

// TestNewMigratorFromFS tests creating a migrator from embed.FS
func TestNewMigratorFromFS(t *testing.T) {
	migrator := NewMigratorFromFS(testMigrations, "testdata")
	if migrator == nil {
		t.Fatal("Expected migrator to be created, got nil")
	}
	if migrator.source == nil {
		t.Fatal("Expected migrator source to be set, got nil")
	}
}

// TestNewMigratorFromDir tests creating a migrator from directory
func TestNewMigratorFromDir(t *testing.T) {
	migrator := NewMigratorFromDir("./testdata")
	if migrator == nil {
		t.Fatal("Expected migrator to be created, got nil")
	}
	if migrator.source == nil {
		t.Fatal("Expected migrator source to be set, got nil")
	}
}

// TestMigratorRun tests running migrations
func TestMigratorRun(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	migrator := NewMigratorFromFS(testMigrations, "testdata")

	result, err := migrator.Run(db)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	if result.Applied != 2 {
		t.Errorf("Expected 2 migrations applied, got %d", result.Applied)
	}

	if result.Skipped != 0 {
		t.Errorf("Expected 0 migrations skipped, got %d", result.Skipped)
	}

	if result.CurrentVersion != "001" {
		t.Errorf("Expected current version to be '001', got '%s'", result.CurrentVersion)
	}
}

// TestMigratorRunIdempotent tests that running migrations twice is idempotent
func TestMigratorRunIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	migrator := NewMigratorFromFS(testMigrations, "testdata")

	// First run
	result1, err := migrator.Run(db)
	if err != nil {
		t.Fatalf("First migration failed: %v", err)
	}

	if result1.Applied != 2 {
		t.Errorf("First run: expected 2 migrations applied, got %d", result1.Applied)
	}

	// Second run - should skip all migrations
	result2, err := migrator.Run(db)
	if err != nil {
		t.Fatalf("Second migration failed: %v", err)
	}

	if result2.Applied != 0 {
		t.Errorf("Second run: expected 0 migrations applied, got %d", result2.Applied)
	}

	if result2.Skipped != 2 {
		t.Errorf("Second run: expected 2 migrations skipped, got %d", result2.Skipped)
	}
}

// TestMigratorStatus tests checking migration status
func TestMigratorStatus(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	migrator := NewMigratorFromFS(testMigrations, "testdata")

	// Check status before any migrations
	status1, err := migrator.Status(db)
	if err != nil {
		t.Fatalf("Status check failed: %v", err)
	}

	if status1.IsUpToDate {
		t.Error("Expected database to not be up to date initially")
	}

	if len(status1.PendingMigrations) != 2 {
		t.Errorf("Expected 2 pending migrations, got %d", len(status1.PendingMigrations))
	}

	if len(status1.AppliedMigrations) != 0 {
		t.Errorf("Expected 0 applied migrations, got %d", len(status1.AppliedMigrations))
	}

	// Run migrations
	_, err = migrator.Run(db)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Check status after migrations
	status2, err := migrator.Status(db)
	if err != nil {
		t.Fatalf("Status check failed: %v", err)
	}

	if !status2.IsUpToDate {
		t.Error("Expected database to be up to date after migrations")
	}

	if len(status2.PendingMigrations) != 0 {
		t.Errorf("Expected 0 pending migrations, got %d", len(status2.PendingMigrations))
	}

	if len(status2.AppliedMigrations) != 2 {
		t.Errorf("Expected 2 applied migrations, got %d", len(status2.AppliedMigrations))
	}

	if status2.CurrentVersion != "001" {
		t.Errorf("Expected current version to be '001', got '%s'", status2.CurrentVersion)
	}
}

// TestGetCurrentVersion tests getting the current database version
func TestGetCurrentVersion(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Before any migrations, version should be empty
	version1, err := GetCurrentVersion(db)
	if err != nil {
		t.Fatalf("GetCurrentVersion failed: %v", err)
	}
	if version1 != "" {
		t.Errorf("Expected empty version before migrations, got '%s'", version1)
	}

	// Run migrations
	migrator := NewMigratorFromFS(testMigrations, "testdata")
	_, err = migrator.Run(db)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// After migrations, version should be the latest
	version2, err := GetCurrentVersion(db)
	if err != nil {
		t.Fatalf("GetCurrentVersion failed: %v", err)
	}
	if version2 != "001" {
		t.Errorf("Expected version '001' after migrations, got '%s'", version2)
	}
}

// TestMigratorWithOptions tests migration with various options
func TestMigratorWithOptions(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	migrator := NewMigratorFromFS(testMigrations, "testdata")

	// Test with verbose option (should not affect result)
	result, err := migrator.Run(db, WithVerbose())
	if err != nil {
		t.Fatalf("Migration with verbose failed: %v", err)
	}

	if result.Applied != 2 {
		t.Errorf("Expected 2 migrations applied, got %d", result.Applied)
	}
}

// TestMigratorDryRun tests dry-run mode
func TestMigratorDryRun(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	migrator := NewMigratorFromFS(testMigrations, "testdata")

	// Run in dry-run mode
	result, err := migrator.Run(db, WithDryRun())
	if err != nil {
		t.Fatalf("Dry-run migration failed: %v", err)
	}

	if result.Applied != 2 {
		t.Errorf("Dry-run: expected 2 migrations to be counted, got %d", result.Applied)
	}

	// In dry-run mode, nothing should actually be applied
	// So version should still be empty (table doesn't exist)
	version, err := GetCurrentVersion(db)
	if err != nil {
		// This is expected - table doesn't exist in dry-run mode
		if version != "" {
			t.Errorf("Expected empty version in dry-run, got '%s'", version)
		}
	} else {
		// If no error, version should be empty
		if version != "" {
			t.Errorf("Expected empty version after dry-run, got '%s'", version)
		}
	}
}

// TestMigrationSourceInterface tests custom migration source
func TestMigrationSourceInterface(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a custom source
	customSource := &testMigrationSource{
		migrations: []Migration{
			{
				ID:          "000",
				Description: "test migration",
				Content: `
					CREATE TABLE IF NOT EXISTS schema_migrations (
						id TEXT PRIMARY KEY,
						description TEXT NOT NULL,
						applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
					);
					CREATE TABLE test (id INTEGER PRIMARY KEY);
				`,
			},
		},
	}

	migrator := NewMigrator(customSource)
	result, err := migrator.Run(db)
	if err != nil {
		t.Fatalf("Custom source migration failed: %v", err)
	}

	if result.Applied != 1 {
		t.Errorf("Expected 1 migration applied, got %d", result.Applied)
	}

	// Verify table was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
	if err != nil {
		t.Errorf("Table 'test' should exist: %v", err)
	}
}

// testMigrationSource is a custom migration source for testing
type testMigrationSource struct {
	migrations []Migration
}

func (s *testMigrationSource) ReadMigrations() ([]Migration, error) {
	return s.migrations, nil
}
