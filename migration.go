// Package migration provides database schema migration functionality for CleanSales-Go-Core.
//
// This package allows applications to programmatically manage database schema migrations
// instead of using command-line tools. It tracks migration history, manages versions,
// and provides APIs to query migration status.
//
// Basic usage:
//
//	import "github.com/gurumfj/cleansales-go-core/migration"
//
//	// Run all pending migrations (using embedded default migrations)
//	result, err := migration.Run(db)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use custom migrations from your own embed.FS
//	//go:embed migrations/*.sql
//	var myMigrations embed.FS
//
//	migrator := migration.NewMigratorFromFS(myMigrations, "migrations")
//	result, err := migrator.Run(db)
//
//	// Use migrations from a directory path
//	migrator := migration.NewMigratorFromDir("./db/migrations")
//	result, err := migrator.Run(db)
//
// Check current version:
//
//	version, err := migration.GetCurrentVersion(db)
package migration

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	// MigrationsTable stores migration history
	MigrationsTable = "schema_migrations"
)

// Migration represents a database migration
type Migration struct {
	ID          string     // Migration ID (e.g., "001")
	Description string     // Human-readable description
	Content     string     // SQL content
	AppliedAt   *time.Time // When it was applied (nil if pending)
}

// MigrationResult contains the result of running migrations
type MigrationResult struct {
	Applied        int       // Number of migrations applied
	Skipped        int       // Number of migrations skipped (already applied)
	CurrentVersion string    // Current database version after migrations
	AppliedAt      time.Time // When migrations were completed
}

// MigrationStatus contains the current migration status
type MigrationStatus struct {
	CurrentVersion    string      // Current database version
	PendingMigrations []Migration // List of pending migrations
	AppliedMigrations []Migration // List of applied migrations
	IsUpToDate        bool        // Whether all migrations are applied
}

// Option configures migration behavior
type Option func(*config)

type config struct {
	dryRun  bool
	verbose bool
	ctx     context.Context
}

// WithDryRun enables dry-run mode (doesn't apply migrations)
func WithDryRun() Option {
	return func(c *config) {
		c.dryRun = true
	}
}

// WithVerbose enables verbose logging
func WithVerbose() Option {
	return func(c *config) {
		c.verbose = true
	}
}

// WithContext sets the context for migration operations
func WithContext(ctx context.Context) Option {
	return func(c *config) {
		c.ctx = ctx
	}
}

// Migrator manages database migrations from a specific source
type Migrator struct {
	source MigrationSource
}

// MigrationSource defines how migrations are loaded
type MigrationSource interface {
	ReadMigrations() ([]Migration, error)
}

// embeddedSource reads migrations from an embed.FS
type embeddedSource struct {
	fs      fs.FS
	rootDir string
}

func (s *embeddedSource) ReadMigrations() ([]Migration, error) {
	return readMigrationsFromFS(s.fs, s.rootDir)
}

// dirSource reads migrations from a filesystem directory
type dirSource struct {
	path string
}

func (s *dirSource) ReadMigrations() ([]Migration, error) {
	return readMigrationsFromDir(s.path)
}

// NewMigrator creates a new Migrator with a custom migration source
func NewMigrator(source MigrationSource) *Migrator {
	return &Migrator{source: source}
}

// NewMigratorFromFS creates a Migrator that reads from an embed.FS
//
// Example:
//
//	//go:embed migrations/*.sql
//	var myMigrations embed.FS
//	migrator := migration.NewMigratorFromFS(myMigrations, "migrations")
func NewMigratorFromFS(fsys fs.FS, rootDir string) *Migrator {
	return &Migrator{
		source: &embeddedSource{
			fs:      fsys,
			rootDir: rootDir,
		},
	}
}

// NewMigratorFromDir creates a Migrator that reads from a filesystem directory
//
// Example:
//
//	migrator := migration.NewMigratorFromDir("./db/migrations")
func NewMigratorFromDir(path string) *Migrator {
	return &Migrator{
		source: &dirSource{path: path},
	}
}

// Run executes all pending migrations for this migrator
func (m *Migrator) Run(db *sql.DB, options ...Option) (*MigrationResult, error) {
	cfg := &config{
		ctx: context.Background(),
	}
	for _, opt := range options {
		opt(cfg)
	}

	// Get all migrations from source
	allMigrations, err := m.source.ReadMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations: %w", err)
	}

	// Get applied migrations (ignore error if table doesn't exist yet)
	appliedMap, err := getAppliedMigrations(db)
	if err != nil && !strings.Contains(err.Error(), "no such table") {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Mark applied migrations
	for i := range allMigrations {
		if appliedAt, ok := appliedMap[allMigrations[i].ID]; ok {
			allMigrations[i].AppliedAt = &appliedAt
		}
	}

	// Apply pending migrations
	applied := 0
	skipped := 0

	for _, migration := range allMigrations {
		if migration.AppliedAt != nil {
			skipped++
			continue
		}

		if cfg.verbose {
			fmt.Printf("Applying migration %s: %s\n", migration.ID, migration.Description)
		}

		if cfg.dryRun {
			fmt.Printf("[DRY RUN] Would apply: %s\n", migration.ID)
			applied++
			continue
		}

		if err := applyMigration(cfg.ctx, db, migration); err != nil {
			return nil, fmt.Errorf("failed to apply migration %s: %w", migration.ID, err)
		}

		applied++
	}

	// Get current version
	currentVersion, err := GetCurrentVersion(db)
	if err != nil {
		// If can't get version, use the latest migration ID as fallback
		if len(allMigrations) > 0 {
			currentVersion = allMigrations[len(allMigrations)-1].ID
		}
	}

	result := &MigrationResult{
		Applied:        applied,
		Skipped:        skipped,
		CurrentVersion: currentVersion,
		AppliedAt:      time.Now(),
	}

	return result, nil
}

// Status returns the current migration status for this migrator
func (m *Migrator) Status(db *sql.DB) (*MigrationStatus, error) {
	// Get all migrations from source
	allMigrations, err := m.source.ReadMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations: %w", err)
	}

	// Get applied migrations (ignore error if table doesn't exist yet)
	appliedMap, err := getAppliedMigrations(db)
	if err != nil && !strings.Contains(err.Error(), "no such table") {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Separate pending and applied migrations
	var pending []Migration
	var applied []Migration

	for _, migration := range allMigrations {
		if appliedAt, ok := appliedMap[migration.ID]; ok {
			migration.AppliedAt = &appliedAt
			applied = append(applied, migration)
		} else {
			pending = append(pending, migration)
		}
	}

	// Get current version
	currentVersion, _ := GetCurrentVersion(db)

	status := &MigrationStatus{
		CurrentVersion:    currentVersion,
		PendingMigrations: pending,
		AppliedMigrations: applied,
		IsUpToDate:        len(pending) == 0,
	}

	return status, nil
}

// readMigrationsFromFS reads all migration files from an fs.FS
func readMigrationsFromFS(fsys fs.FS, rootDir string) ([]Migration, error) {
	var migrations []Migration

	err := fs.WalkDir(fsys, rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		basename := filepath.Base(path)

		// Parse filename: 000_baseline_v0_3_0.sql
		parts := strings.SplitN(basename, "_", 2)
		if len(parts) != 2 {
			return nil // Skip files that don't match pattern
		}

		id := parts[0]
		description := strings.TrimSuffix(parts[1], ".sql")
		description = strings.ReplaceAll(description, "_", " ")

		// Read content
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		migrations = append(migrations, Migration{
			ID:          id,
			Description: description,
			Content:     string(content),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by ID
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	return migrations, nil
}

// readMigrationsFromDir reads all migration files from a filesystem directory
func readMigrationsFromDir(dirPath string) ([]Migration, error) {
	var migrations []Migration

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		basename := filepath.Base(path)

		// Parse filename: 000_baseline_v0_3_0.sql
		parts := strings.SplitN(basename, "_", 2)
		if len(parts) != 2 {
			return nil // Skip files that don't match pattern
		}

		id := parts[0]
		description := strings.TrimSuffix(parts[1], ".sql")
		description = strings.ReplaceAll(description, "_", " ")

		// Read content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		migrations = append(migrations, Migration{
			ID:          id,
			Description: description,
			Content:     string(content),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by ID
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	return migrations, nil
}

// getAppliedMigrations retrieves the list of already applied migrations
func getAppliedMigrations(db *sql.DB) (map[string]time.Time, error) {
	query := fmt.Sprintf(`
		SELECT id, applied_at
		FROM %s
		ORDER BY id
	`, MigrationsTable)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]time.Time)
	for rows.Next() {
		var id string
		var appliedAt time.Time
		if err := rows.Scan(&id, &appliedAt); err != nil {
			return nil, err
		}
		applied[id] = appliedAt
	}

	return applied, rows.Err()
}

// applyMigration executes a migration and records it as applied
func applyMigration(ctx context.Context, db *sql.DB, migration Migration) error {
	// Execute migration in a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.ExecContext(ctx, migration.Content); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied
	query := fmt.Sprintf(`
		INSERT INTO %s (id, description)
		VALUES (?, ?)
	`, MigrationsTable)

	if _, err := tx.ExecContext(ctx, query, migration.ID, migration.Description); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetCurrentVersion returns the current highest migration ID that has been applied
// Returns empty string if schema_migrations table doesn't exist or no migrations have been applied
//
// Examples:
//   - Returns "001" if migrations 000 and 001 have been applied
//   - Returns "" if no migrations have been applied yet
func GetCurrentVersion(db *sql.DB) (string, error) {
	query := "SELECT MAX(id) FROM schema_migrations"

	var version sql.NullString
	err := db.QueryRow(query).Scan(&version)
	if err != nil {
		// Handle table not existing (before migration 000 runs)
		if strings.Contains(err.Error(), "no such table") {
			return "", nil
		}
		return "", err
	}

	if !version.Valid {
		return "", nil // Table exists but no records
	}

	return version.String, nil
}
