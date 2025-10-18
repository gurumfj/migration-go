package migration

import (
	"database/sql"
	"strings"
)

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
