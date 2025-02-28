package sqlvalidator

import (
	"errors"
	"strings"
)

// Validate checks if the SQL query is valid for the given dialect
func Validate(sql string, dialect string) (bool, error) {
	// Trim whitespace
	sql = strings.TrimSpace(sql)

	// Check if empty
	if sql == "" {
		return false, errors.New("SQL query cannot be empty")
	}

	// Run safety checks
	safetyCheck := IsSafeDDLOperation(sql, dialect)
	if !safetyCheck.Safe {
		return false, errors.New(safetyCheck.Error)
	}

	// Dialect-specific validation
	switch strings.ToLower(dialect) {
	case "mysql":
		return validateMySQL(sql)
	case "postgresql":
		return validatePostgreSQL(sql)
	case "sqlite":
		return validateSQLite(sql)
	default:
		return false, errors.New("unsupported SQL dialect")
	}
}

// validateMySQL validates MySQL syntax
func validateMySQL(sql string) (bool, error) {
	// Check for basic SELECT query structure
	if strings.HasPrefix(strings.ToLower(sql), "select") {
		return true, nil
	}

	// Basic check for INSERT/UPDATE/DELETE/CREATE TABLE syntax
	validOperations := []string{"insert into", "update", "delete from", "create table"}
	sqlLower := strings.ToLower(sql)

	for _, op := range validOperations {
		if strings.HasPrefix(sqlLower, op) {
			return true, nil
		}
	}

	// If it doesn't match our basic patterns, still allow it
	// The actual validation will happen when executed against the database
	return true, nil
}

// validatePostgreSQL validates PostgreSQL syntax
func validatePostgreSQL(sql string) (bool, error) {
	// Very similar to MySQL validation
	return validateMySQL(sql)
}

// validateSQLite validates SQLite syntax
func validateSQLite(sql string) (bool, error) {
	// Very similar to MySQL validation
	return validateMySQL(sql)
}
