# Migration-Go Project Overview

## Purpose
A flexible, easy-to-use database migration library for Go applications. This library allows developers to programmatically manage database schema migrations with support for embedded migrations, filesystem migrations, and custom migration sources.

## Tech Stack
- **Language**: Go 1.25.2
- **Database Driver**: SQLite (github.com/mattn/go-sqlite3 v1.14.32)
- **Standard Library**: Uses Go's `embed`, `database/sql`, and `io/fs` packages

## Key Features
- 📦 Embedded Migrations: Embed SQL files directly into binary using Go's embed package
- 📁 Filesystem Migrations: Load migrations from a directory at runtime
- 🔌 Extensible: Implement custom migration sources
- 📊 Migration Tracking: Automatic tracking of applied migrations in schema_migrations table
- 🔄 Transaction Support: Each migration runs in a transaction for safety
- ✅ Status Checking: Query current version and pending migrations

## Project Structure
```
migration-go/
├── migration.go          # Core migration library implementation
├── version.go           # Helper function to get current version
├── example_test.go      # Example tests demonstrating usage
├── testdata/            # Test migration files
│   ├── 000_test_baseline.sql
│   └── 001_add_user_profile.sql
├── examples/            # Example projects
│   └── embedded/        # Example of embedded migrations
│       ├── main.go
│       ├── README.md
│       └── migrations/
├── README.md            # Main documentation
├── go.mod
└── go.sum
```

## Migration File Naming Convention
Files must follow: `{ID}_{description}.sql`
- ID: Zero-padded sequential number (000, 001, 002...)
- Description: Lowercase with underscores
- Example: `000_initial_schema.sql`, `001_add_users_table.sql`
