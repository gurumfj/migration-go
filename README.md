# Migration - Database Migration Library for Go

A flexible, easy-to-use database migration library for Go applications. This library allows you to programmatically manage database schema migrations with support for embedded migrations, filesystem migrations, and custom migration sources.

## Features

- 📦 **Embedded Migrations**: Embed SQL files directly into your binary using Go's `embed` package
- 📁 **Filesystem Migrations**: Load migrations from a directory at runtime
- 🔌 **Extensible**: Implement custom migration sources for advanced use cases
- 📊 **Migration Tracking**: Automatic tracking of applied migrations
- 🔄 **Transaction Support**: Each migration runs in a transaction for safety
- 🎯 **Simple API**: Clean, intuitive API with sensible defaults
- ✅ **Status Checking**: Query current version and pending migrations
- 🔧 **Options**: Dry-run mode, verbose logging, custom context support

## Installation

```bash
go get migration-go
```

## Quick Start

### Using Embedded Migrations

```go
import (
    "database/sql"
    "log"
    
    migration "migration-go"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    db, err := sql.Open("sqlite3", "./app.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Run all pending migrations
    result, err := migration.Run(db)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Applied %d migrations, current version: %s", 
        result.Applied, result.CurrentVersion)
}
```

## Using in Your Own Projects

The real power of this library comes from using it in your own projects with custom migrations.

### Method 1: Embedded Migrations (Recommended)

Create your migration files and embed them in your application:

**Project Structure:**
```
your-project/
├── main.go
├── migrations/
│   ├── 000_initial_schema.sql
│   ├── 001_add_users_table.sql
│   └── 002_add_posts_table.sql
└── go.mod
```

**main.go:**
```go
package main

import (
    "database/sql"
    "embed"
    "log"
    
    migration "migration-go"
    _ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
    db, err := sql.Open("sqlite3", "./myapp.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create migrator with your embedded migrations
    migrator := migration.NewMigratorFromFS(migrations, "migrations")
    
    // Run migrations
    result, err := migrator.Run(db, migration.WithVerbose())
    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }

    log.Printf("✅ Applied %d migrations", result.Applied)
    log.Printf("📊 Current version: %s", result.CurrentVersion)
}
```

### Method 2: Filesystem Migrations

Load migrations from a directory at runtime (useful for development):

```go
package main

import (
    "database/sql"
    "log"
    
    migration "migration-go"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    db, _ := sql.Open("sqlite3", "./myapp.db")
    defer db.Close()

    // Create migrator that reads from filesystem
    migrator := migration.NewMigratorFromDir("./db/migrations")
    
    // Run migrations
    result, err := migrator.Run(db)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Applied %d migrations", result.Applied)
}
```

### Method 3: Custom Migration Source

For advanced use cases, implement the `MigrationSource` interface:

```go
package main

import (
    migration "migration-go"
)

// CustomSource loads migrations from your custom location
// (e.g., database, HTTP API, encrypted files, etc.)
type CustomSource struct {
    apiURL string
}

func (s *CustomSource) ReadMigrations() ([]migration.Migration, error) {
    // Your custom logic here
    // Fetch migrations from API, decrypt files, etc.
    
    return []migration.Migration{
        {
            ID:          "000",
            Description: "initial schema",
            Content:     "CREATE TABLE ...",
        },
    }, nil
}

func main() {
    db, _ := sql.Open("sqlite3", "./myapp.db")
    defer db.Close()

    // Use custom source
    source := &CustomSource{apiURL: "https://api.example.com/migrations"}
    migrator := migration.NewMigrator(source)
    
    result, err := migrator.Run(db)
    // ...
}
```

## Migration File Format

Migration files must follow this naming convention:

```
{ID}_{description}.sql
```

- **ID**: Zero-padded number (e.g., `000`, `001`, `002`)
- **Description**: Underscore-separated description

**Examples:**
- `000_initial_schema.sql`
- `001_add_users_table.sql`
- `002_add_email_index.sql`

**Sample Migration File:**
```sql
-- 001_add_users_table.sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

## API Reference

### Package-Level Functions (Default Migrations)

These use the library's embedded default migrations:

```go
// Run all pending migrations
result, err := migration.Run(db, options...)

// Get current migration status
status, err := migration.Status(db)

// Get current version
version, err := migration.GetCurrentVersion(db)
```

### Migrator Methods (Custom Migrations)

For custom migrations, create a `Migrator` first:

```go
// Create from embed.FS
migrator := migration.NewMigratorFromFS(fsys, "migrations")

// Create from directory
migrator := migration.NewMigratorFromDir("./db/migrations")

// Create from custom source
migrator := migration.NewMigrator(customSource)

// Run migrations
result, err := migrator.Run(db, options...)

// Check status
status, err := migrator.Status(db)
```

### Options

```go
// Dry-run mode (shows what would be applied without actually applying)
migration.WithDryRun()

// Verbose logging
migration.WithVerbose()

// Custom context
migration.WithContext(ctx)
```

**Example:**
```go
result, err := migrator.Run(db, 
    migration.WithVerbose(),
    migration.WithContext(ctx),
)
```

### Types

#### MigrationResult
```go
type MigrationResult struct {
    Applied        int       // Number of migrations applied
    Skipped        int       // Number already applied
    CurrentVersion string    // Current database version
    AppliedAt      time.Time // When completed
}
```

#### MigrationStatus
```go
type MigrationStatus struct {
    CurrentVersion    string      // Current version
    PendingMigrations []Migration // Pending migrations
    AppliedMigrations []Migration // Applied migrations
    IsUpToDate        bool        // All migrations applied?
}
```

#### Migration
```go
type Migration struct {
    ID          string     // Migration ID (e.g., "001")
    Description string     // Human-readable description
    Content     string     // SQL content
    AppliedAt   *time.Time // When applied (nil if pending)
}
```

## Common Use Cases

### Application Startup Migration

```go
func main() {
    db := setupDatabase()
    defer db.Close()

    // Run migrations automatically on startup
    log.Println("Running database migrations...")
    migrator := migration.NewMigratorFromFS(migrations, "migrations")
    result, err := migrator.Run(db)
    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }

    if result.Applied > 0 {
        log.Printf("✅ Applied %d new migrations", result.Applied)
    } else {
        log.Println("✅ Database is up to date")
    }

    // Start application...
}
```

### Check Migration Status

```go
func checkMigrationStatus(db *sql.DB) {
    migrator := migration.NewMigratorFromFS(migrations, "migrations")
    status, err := migrator.Status(db)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Current Version: %s\n", status.CurrentVersion)
    fmt.Printf("Up to Date: %v\n", status.IsUpToDate)
    
    if len(status.PendingMigrations) > 0 {
        fmt.Println("\nPending Migrations:")
        for _, m := range status.PendingMigrations {
            fmt.Printf("  - %s: %s\n", m.ID, m.Description)
        }
    }
}
```

### Dry Run Before Applying

```go
func safeMigration(db *sql.DB) {
    migrator := migration.NewMigratorFromFS(migrations, "migrations")
    
    // First, do a dry run to see what would be applied
    result, err := migrator.Run(db, 
        migration.WithDryRun(), 
        migration.WithVerbose(),
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Would apply %d migrations\n", result.Applied)
    
    // Prompt user for confirmation...
    if userConfirms() {
        // Actually run the migrations
        result, err = migrator.Run(db, migration.WithVerbose())
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Applied %d migrations successfully\n", result.Applied)
    }
}
```

### Multiple Databases

```go
func migrateMultipleDatabases() {
    // Different projects can use different migration sources
    
    // Project A
    dbA, _ := sql.Open("sqlite3", "./projectA.db")
    migratorA := migration.NewMigratorFromFS(migrationsA, "migrations")
    migratorA.Run(dbA)
    
    // Project B
    dbB, _ := sql.Open("postgres", "postgres://...")
    migratorB := migration.NewMigratorFromDir("./project-b/migrations")
    migratorB.Run(dbB)
}
```

## Migration Strategy

The library uses a simple, linear migration strategy:

1. **Sequential IDs**: Migrations are numbered sequentially (000, 001, 002, ...)
2. **Baseline Migration (000)**: Creates the `schema_migrations` table using `IF NOT EXISTS`
3. **Incremental Changes**: Each subsequent migration builds on the previous state
4. **Transaction Safety**: Each migration runs in its own transaction
5. **Idempotent**: Safe to run multiple times (already-applied migrations are skipped)

## Database Support

This library works with any SQL database supported by Go's `database/sql` package:

- ✅ SQLite
- ✅ PostgreSQL
- ✅ MySQL/MariaDB
- ✅ SQL Server
- ✅ Oracle
- ✅ CockroachDB

Just import the appropriate driver:

```go
_ "github.com/mattn/go-sqlite3"        // SQLite
_ "github.com/lib/pq"                   // PostgreSQL
_ "github.com/go-sql-driver/mysql"      // MySQL
_ "github.com/denisenkom/go-mssqldb"    // SQL Server
```

## Best Practices

1. **Always Use Transactions**: The library automatically wraps each migration in a transaction
2. **Keep Migrations Small**: One logical change per migration
3. **Never Modify Applied Migrations**: Once a migration is applied, create a new one for changes
4. **Use Descriptive Names**: `001_add_user_table.sql` is better than `001_update.sql`
5. **Test Migrations**: Test on a copy of production data before applying to production
6. **Version Control**: Commit all migration files to version control
7. **Sequential IDs**: Always use the next sequential number for new migrations

## Troubleshooting

### Migration Fails

If a migration fails:
1. The transaction is rolled back automatically
2. No migration record is created
3. The error is returned to the caller
4. Fix the migration SQL and try again

### Schema Migrations Table

The library creates a `schema_migrations` table to track applied migrations:

```sql
CREATE TABLE schema_migrations (
    id TEXT PRIMARY KEY,
    description TEXT NOT NULL,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

You can query this table directly to see migration history:

```sql
SELECT * FROM schema_migrations ORDER BY id;
```

## Examples

See the [examples directory](./examples/) for complete working examples:

- **[Basic Usage](./examples/basic/)** - Comprehensive examples demonstrating all library features
  - Basic migration usage
  - Status checking
  - Verbose output
  - Version tracking
  - Filesystem migrations
  
- **[Embedded Migrations](./examples/embedded/)** - Production-ready application with embedded migrations
  - Complete startup flow
  - Migration status checking
  - Error handling
  - Database connection management

Each example includes:
- Complete, runnable code
- Detailed README with explanations
- Sample migration files
- Expected output

## Contributing

Contributions are welcome! Please ensure:
- Tests pass
- Code is formatted with `go fmt`
- Documentation is updated
- Examples are provided for new features

## License

MIT License