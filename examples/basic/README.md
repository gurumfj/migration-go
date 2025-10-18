# Basic Usage Examples

This directory contains comprehensive examples demonstrating various usage patterns of the migration library.

## What's Inside

- `main.go` - Executable program with multiple usage examples
- `migrations/` - Sample migration files
  - `000_test_baseline.sql` - Initial schema setup
  - `001_add_user_profile.sql` - User profile feature addition

## Running the Examples

```bash
# Navigate to this directory
cd examples/basic

# Run all examples
go run main.go
```

## Examples Included

### Example 1: Basic Migration
The simplest way to use the library - create a migrator and run migrations.

```go
migrator := migration.NewMigratorFromFS(migrations, "migrations")
result, err := migrator.Run(db)
```

### Example 2: Check Migration Status
Check what migrations are pending before applying them.

```go
status, err := migrator.Status(db)
fmt.Printf("Pending: %d, Applied: %d\n", 
    len(status.PendingMigrations), 
    len(status.AppliedMigrations))
```

### Example 3: Verbose Output
Enable detailed logging to see what's happening during migration.

```go
result, err := migrator.Run(db, migration.WithVerbose())
```

### Example 4: Get Current Version
Query the current database version and migration history.

```go
version, err := migration.GetCurrentVersion(db)
```

### Example 5: Filesystem Migrations
Load migrations from a directory instead of embedding them.

```go
migrator := migration.NewMigratorFromDir("./migrations")
```

## Expected Output

When you run the examples, you should see:

```
🚀 Migration Library - Basic Usage Examples

=============================================================
Example 1: Basic Migration
=============================================================
✓ Applied 2 migrations
✓ Current version: 001

=============================================================
Example 2: Check Migration Status
=============================================================
Current version: 
Is up to date: false
Pending migrations: 2
Applied migrations: 0

Pending migrations:
  - 000: test baseline
  - 001: add user profile

✓ Applied 2 migrations

=============================================================
Example 3: Verbose Migration Output
=============================================================
Applying migration 000: test baseline
Applying migration 001: add user profile

✓ Migration completed: 2 applied

=============================================================
Example 4: Get Current Version
=============================================================
✓ Current migration ID: 001

Migration history:
  - [000] test baseline (applied at: 2024-10-18 14:30:00)
  - [001] add user profile (applied at: 2024-10-18 14:30:00)

=============================================================
Example 5: Filesystem Migrations
=============================================================
✓ Applied 2 migrations from filesystem

✅ All examples completed successfully!
```

## Key Takeaways

1. **Embedded Migrations** (Examples 1-4): Best for production - migrations are bundled in your binary
2. **Filesystem Migrations** (Example 5): Best for development - easy to modify migrations
3. **Always Check Status**: Use `Status()` to see what will be applied before running
4. **Verbose Mode**: Helpful for debugging and understanding migration flow
5. **Version Tracking**: Use `GetCurrentVersion()` to query current state

## Using in Your Project

To use these patterns in your own project:

1. **Copy the structure**:
   ```bash
   mkdir -p myproject/migrations
   ```

2. **Add embed directive** to your main.go:
   ```go
   //go:embed migrations/*.sql
   var migrations embed.FS
   ```

3. **Create your migrations** following the naming convention:
   ```
   000_initial_schema.sql
   001_add_users.sql
   002_add_posts.sql
   ```

4. **Run migrations on startup**:
   ```go
   migrator := migration.NewMigratorFromFS(migrations, "migrations")
   result, err := migrator.Run(db)
   ```

## Learn More

- See [embedded example](../embedded/) for a complete production-ready application
- Read the [main README](../../README.md) for full API documentation
- Check [migration.go](../../migration.go) for implementation details