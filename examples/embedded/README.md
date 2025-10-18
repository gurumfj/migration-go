# Embedded Migrations Example

This example demonstrates how to use the `migration` library in your own Go project with embedded SQL migration files.

## Project Structure

```
embedded/
├── main.go                          # Application entry point
├── migrations/                      # Migration files directory
│   ├── 000_initial_schema.sql      # Initial database schema
│   └── 001_add_comments.sql        # Add comments feature
├── go.mod                           # Go module file (you need to create this)
└── README.md                        # This file
```

## How It Works

1. **Embed Migrations**: The `//go:embed` directive embeds all `.sql` files from the `migrations/` directory into your binary at compile time.

2. **Create Migrator**: Use `migration.NewMigratorFromFS()` to create a migrator that reads from your embedded filesystem.

3. **Run Migrations**: Call `migrator.Run()` to execute all pending migrations automatically.

## Running This Example

### Prerequisites

```bash
# Make sure you have Go 1.16+ installed
go version

# Install SQLite driver
go get github.com/mattn/go-sqlite3
```

### Initialize Go Module (if not already done)

```bash
cd embedded
go mod init myapp
go mod tidy
```

### Run the Example

```bash
# First run - will create database and apply all migrations
go run main.go

# Second run - will skip already-applied migrations
go run main.go
```

### Expected Output

First run:
```
🚀 Starting application with embedded migrations...
📊 Checking migration status...
Current version: 
Pending migrations: 2
Applied migrations: 0
🔄 Running pending migrations...
Applying migration 000: initial schema
Applying migration 001: add comments
✅ Applied 2 migrations
📋 Current version: 001
🎉 Application started successfully!
Users table has 0 records
```

Second run:
```
🚀 Starting application with embedded migrations...
📊 Checking migration status...
Current version: 001
Pending migrations: 0
Applied migrations: 2
✅ Database is already up to date!
🎉 Application started successfully!
Users table has 0 records
```

## Key Code Sections

### Embedding Migrations

```go
//go:embed migrations/*.sql
var migrations embed.FS
```

This directive tells Go to embed all `.sql` files from the `migrations/` directory into your binary.

### Creating the Migrator

```go
migrator := migration.NewMigratorFromFS(migrations, "migrations")
```

Creates a migrator that will read migrations from the embedded filesystem, starting from the `"migrations"` directory.

### Running Migrations

```go
result, err := migrator.Run(db, migration.WithVerbose())
if err != nil {
    log.Fatalf("Migration failed: %v", err)
}
```

Executes all pending migrations in order.

## Migration File Naming Convention

Migration files must follow this pattern:

```
{ID}_{description}.sql
```

- **ID**: Zero-padded sequential number (000, 001, 002, ...)
- **Description**: Lowercase with underscores

Examples:
- ✅ `000_initial_schema.sql`
- ✅ `001_add_comments.sql`
- ✅ `002_add_user_profiles.sql`
- ❌ `1_init.sql` (ID not zero-padded)
- ❌ `001-add-feature.sql` (use underscores, not hyphens)

## Adding New Migrations

To add a new migration:

1. Create a new SQL file in `migrations/` with the next sequential ID:
   ```bash
   touch migrations/002_add_tags.sql
   ```

2. Write your SQL changes:
   ```sql
   -- 002_add_tags.sql
   CREATE TABLE tags (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       name TEXT NOT NULL UNIQUE
   );
   ```

3. Restart your application - the new migration will be applied automatically:
   ```bash
   go run main.go
   ```

## Benefits of This Approach

✅ **Self-Contained**: Migrations are embedded in your binary - no external files needed in production

✅ **Version Control**: Migration files are committed alongside your code

✅ **Automatic**: Migrations run automatically on application startup

✅ **Safe**: Each migration runs in a transaction - if it fails, changes are rolled back

✅ **Idempotent**: Safe to run multiple times - already-applied migrations are skipped

✅ **Trackable**: Migration history is stored in the `schema_migrations` table

## Customization Options

### Verbose Logging

```go
result, err := migrator.Run(db, migration.WithVerbose())
```

Prints detailed information about each migration as it's applied.

### Dry Run

```go
result, err := migrator.Run(db, migration.WithDryRun(), migration.WithVerbose())
```

Shows what would be applied without actually applying migrations.

### Custom Context

```go
ctx := context.WithTimeout(context.Background(), 30*time.Second)
result, err := migrator.Run(db, migration.WithContext(ctx))
```

Use a custom context for timeouts or cancellation.

## Using in Your Own Project

To use this pattern in your own project:

1. **Copy the structure**:
   ```bash
   mkdir -p myproject/migrations
   ```

2. **Add the embed directive** to your `main.go`:
   ```go
   //go:embed migrations/*.sql
   var migrations embed.FS
   ```

3. **Create your first migration** (`migrations/000_initial_schema.sql`):
   ```sql
   CREATE TABLE IF NOT EXISTS schema_migrations (
       id TEXT PRIMARY KEY,
       description TEXT NOT NULL,
       applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
   );
   
   -- Your tables here...
   ```

4. **Initialize the migrator** in your `main()`:
   ```go
   migrator := migration.NewMigratorFromFS(migrations, "migrations")
   result, err := migrator.Run(db)
   ```

5. **Deploy**: Your binary now contains all migrations and will automatically update the database schema on startup.

## Troubleshooting

### "no such file or directory" error

Make sure:
- Your migration files are in the `migrations/` directory
- You're using the correct path in `NewMigratorFromFS()`: `"migrations"` (without leading slash)
- Your `//go:embed` directive matches the actual directory structure

### Migration fails with SQL error

- Check your SQL syntax in the migration file
- The migration will be rolled back automatically
- Fix the SQL and restart the application
- The migration will be retried

### "table already exists" error

Use `IF NOT EXISTS` in your baseline (000) migration:
```sql
CREATE TABLE IF NOT EXISTS users (...);
```

This makes your baseline migration idempotent.

## Next Steps

- See [filesystem example](../filesystem/) for loading migrations from a directory instead of embedding
- See [custom source example](../custom/) for implementing your own migration source
- Read the [main README](../../README.md) for complete API documentation