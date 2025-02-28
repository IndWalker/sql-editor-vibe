package sqlvalidator

import (
	"regexp"
	"strings"
)

// SafetyCheckResult represents the result of a safety check
type SafetyCheckResult struct {
	Safe  bool
	Error string
}

// IsSafeDDLOperation checks if a Data Definition Language (DDL) operation is safe
func IsSafeDDLOperation(sql string, dialect string) SafetyCheckResult {
	sqlLower := strings.ToLower(sql)

	// Explicitly block dangerous operations
	blockedPatterns := []struct {
		pattern string
		message string
	}{
		{`drop\s+(database|schema|user)`, "DROP DATABASE/SCHEMA/USER operations are not allowed"},
		{`truncate\s+database`, "TRUNCATE DATABASE operations are not allowed"},
		{`delete\s+from\s+(user|users|permission|permissions|role|roles|account|accounts)`, "DELETE operations on sensitive tables are not allowed"},
		{`alter\s+user`, "ALTER USER operations are not allowed"},
		{`grant\s+all`, "GRANT ALL operations are not allowed"},
		{`revoke\s+all`, "REVOKE ALL operations are not allowed"},
		{`shutdown`, "SHUTDOWN operations are not allowed"},
		{`create\s+(database|schema)`, "CREATE DATABASE/SCHEMA operations are not allowed"},
		{`drop\s+table`, "DROP TABLE operations are not allowed in this playground"},
		{`alter\s+table\s+\w+\s+drop\s+column`, "ALTER TABLE DROP COLUMN operations are not allowed"},
		{`delete\s+from\s+\w+\s+where\s+1\s*=\s*1`, "DELETE all records operations are not allowed"},
		{`update\s+\w+\s+set\s+.+where\s+1\s*=\s*1`, "UPDATE all records operations are not allowed"},
		{`(;|--)\s*(drop|delete|update|insert|alter|create)`, "SQL injection attempts are not allowed"},
	}

	for _, blockedPattern := range blockedPatterns {
		matched, err := regexp.MatchString(blockedPattern.pattern, sqlLower)
		if err != nil {
			continue // Skip this pattern if there's a regex error
		}
		if matched {
			return SafetyCheckResult{
				Safe:  false,
				Error: blockedPattern.message,
			}
		}
	}

	// Restrict operations based on dialect
	switch dialect {
	case "sqlite":
		return verifySQLiteSafety(sqlLower)
	case "mysql":
		return verifyMySQLSafety(sqlLower)
	case "postgresql":
		return verifyPostgreSQLSafety(sqlLower)
	default:
		return SafetyCheckResult{
			Safe:  false,
			Error: "Unsupported SQL dialect",
		}
	}
}

// verifySQLiteSafety checks if an operation is safe for SQLite
func verifySQLiteSafety(sqlLower string) SafetyCheckResult {
	// Protect against PRAGMA changes that could modify database behavior
	if strings.Contains(sqlLower, "pragma") &&
		(strings.Contains(sqlLower, "journal_mode") ||
			strings.Contains(sqlLower, "synchronous") ||
			strings.Contains(sqlLower, "secure_delete")) {
		return SafetyCheckResult{
			Safe:  false,
			Error: "PRAGMA statements that modify database settings are not allowed",
		}
	}

	// Restrict attaching other databases
	if strings.Contains(sqlLower, "attach database") {
		return SafetyCheckResult{
			Safe:  false,
			Error: "ATTACH DATABASE operations are not allowed",
		}
	}

	return SafetyCheckResult{Safe: true}
}

// verifyMySQLSafety checks if an operation is safe for MySQL
func verifyMySQLSafety(sqlLower string) SafetyCheckResult {
	// Block system table modifications
	if (strings.Contains(sqlLower, "mysql.") ||
		strings.Contains(sqlLower, "information_schema.") ||
		strings.Contains(sqlLower, "performance_schema.")) &&
		(strings.Contains(sqlLower, "insert") ||
			strings.Contains(sqlLower, "update") ||
			strings.Contains(sqlLower, "delete") ||
			strings.Contains(sqlLower, "alter")) {
		return SafetyCheckResult{
			Safe:  false,
			Error: "Modifying system tables is not allowed",
		}
	}

	// Block system variable changes
	if strings.Contains(sqlLower, "set global") ||
		strings.Contains(sqlLower, "set @@global") {
		return SafetyCheckResult{
			Safe:  false,
			Error: "Setting global variables is not allowed",
		}
	}

	return SafetyCheckResult{Safe: true}
}

// verifyPostgreSQLSafety checks if an operation is safe for PostgreSQL
func verifyPostgreSQLSafety(sqlLower string) SafetyCheckResult {
	// Block system catalog modifications
	if strings.Contains(sqlLower, "pg_") &&
		(strings.Contains(sqlLower, "insert") ||
			strings.Contains(sqlLower, "update") ||
			strings.Contains(sqlLower, "delete") ||
			strings.Contains(sqlLower, "alter")) {
		return SafetyCheckResult{
			Safe:  false,
			Error: "Modifying system catalogs is not allowed",
		}
	}

	// Block dangerous function calls
	dangerousFunctions := []string{
		"pg_read_file", "pg_ls_dir", "pg_sleep",
		"copy", "lo_import", "lo_export",
		"pg_catalog.pg_file_write", "pg_catalog.pg_read_binary_file",
	}

	for _, function := range dangerousFunctions {
		if strings.Contains(sqlLower, function) {
			return SafetyCheckResult{
				Safe:  false,
				Error: "Usage of potentially dangerous functions is not allowed: " + function,
			}
		}
	}

	return SafetyCheckResult{Safe: true}
}

// HasLimitForSelect checks if SELECT statements have a LIMIT clause
// and adds a default limit if necessary
func HasLimitForSelect(sql string) (string, bool) {
	sqlLower := strings.ToLower(sql)

	// If it's not a SELECT statement, no change needed
	if !strings.HasPrefix(strings.TrimSpace(sqlLower), "select") {
		return sql, false
	}

	// Check if LIMIT is already present
	limitRegex := regexp.MustCompile(`\s+limit\s+\d+`)
	if limitRegex.MatchString(sqlLower) {
		return sql, false
	}

	// Add a default LIMIT of 100 rows
	modifiedSQL := sql + " LIMIT 100"
	return modifiedSQL, true
}

// SanitizeIdentifiers ensures that table and column identifiers are properly quoted
func SanitizeIdentifiers(sql string, dialect string) string {
	// This is a simplified version. In reality, this would require a proper SQL parser
	// to correctly identify and quote all identifiers.

	// For the playground purposes, just ensure basic safety
	return sql
}
