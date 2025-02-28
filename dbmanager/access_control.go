package dbmanager

import (
	"context"
	"database/sql"
	"time"
)

// ExecuteWithTimeout executes a SQL query with a specified timeout
// This prevents long-running queries from consuming resources
func ExecuteWithTimeout(db *sql.DB, query string) (*sql.Rows, error) {
	// Create a context with a timeout of 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Execute the query with the timeout context
	return db.QueryContext(ctx, query)
}

// SetSafeDatabaseDefaults ensures safe database settings
func SetSafeDatabaseDefaults(db *sql.DB, dialect string) error {
	switch dialect {
	case "mysql":
		// Set MySQL in safe mode - no file operations allowed
		_, err := db.Exec("SET GLOBAL local_infile=0")
		if err != nil {
			return err
		}
		// Limit max connections per user to prevent resource exhaustion
		_, err = db.Exec("SET GLOBAL max_user_connections=3")
		return err

	case "postgresql":
		// Disable file system access
		_, err := db.Exec("SET lo_compat_privileges = off")
		return err

	case "sqlite":
		// SQLite has fewer runtime configuration options
		// Just ensure foreign keys are enabled for consistency
		_, err := db.Exec("PRAGMA foreign_keys = ON")
		return err
	}

	return nil
}

// ApplyTransactionLimits ensures that transactions have reasonable timeouts
func ApplyTransactionLimits(db *sql.DB, dialect string) error {
	switch dialect {
	case "mysql":
		// Set a 5 second lock wait timeout
		_, err := db.Exec("SET innodb_lock_wait_timeout = 5")
		return err

	case "postgresql":
		// Set a 5 second statement timeout
		_, err := db.Exec("SET statement_timeout = 5000") // 5000ms = 5s
		return err
	}

	return nil
}
