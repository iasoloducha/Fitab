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

// migrate runs all SQL migrations in order
func (db *DB) migrate() error {
	// Read embedded migration directory listing
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("listing migrations: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		migrationFiles, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", entry.Name(), err)
		}

		// Split by semicolons and execute each statement
		statements := strings.Split(string(migrationFiles), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				return fmt.Errorf("executing migration %s: %w\nStatement: %s", entry.Name(), err, stmt)
			}
		}
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}