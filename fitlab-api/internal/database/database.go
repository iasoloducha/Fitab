package database

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// FS is used to embed migrations
//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB holds the database connection
type DB struct {
	*sql.DB
	Path string
}

// New creates a new database connection and runs migrations
func New(dbPath string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("creating db directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	wrapper := &DB{DB: db, Path: dbPath}
	if err := wrapper.migrate(); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return wrapper, nil
}

// columnExists checks if a column exists in a table
func (db *DB) columnExists(table, column string) (bool, error) {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue *string
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

// migrate runs all SQL migrations in order, tracking which have been applied
func (db *DB) migrate() error {
	// Create migration tracking table if it doesn't exist
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS _migrations (
		name TEXT PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	// Bootstrap: detect already-applied migrations for existing databases
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM _migrations").Scan(&count); err != nil {
		return fmt.Errorf("checking migrations count: %w", err)
	}

	if count == 0 {
		// Detect migration 002: check if actual_sets exists in exercise_logs
		if applied, err := db.columnExists("exercise_logs", "actual_sets"); err != nil {
			return fmt.Errorf("checking migration 002 state: %w", err)
		} else if applied {
			if _, err := db.Exec("INSERT INTO _migrations (name) VALUES ('002_add_sets_reps_to_logs.sql')"); err != nil {
				return fmt.Errorf("recording existing migration 002: %w", err)
			}
		}
	}

	// Read embedded migration directory listing
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("listing migrations: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		migrationName := entry.Name()

		// Check if this migration was already applied
		var exists bool
		if err := db.QueryRow("SELECT 1 FROM _migrations WHERE name = ?", migrationName).Scan(&exists); err == nil {
			continue // already applied
		}

		migrationFile, err := migrationsFS.ReadFile("migrations/" + migrationName)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", migrationName, err)
		}

		// Split by semicolons and execute each statement
		statements := strings.Split(string(migrationFile), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				return fmt.Errorf("executing migration %s: %w\nStatement: %s", migrationName, err, stmt)
			}
		}

		// Record migration as applied
		if _, err := db.Exec("INSERT OR IGNORE INTO _migrations (name) VALUES (?)", migrationName); err != nil {
			return fmt.Errorf("recording migration %s: %w", migrationName, err)
		}
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}