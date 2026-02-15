// Package storage provides state persistence for migrations.
package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// MigrationState represents the state of a migration
type MigrationState struct {
	MigrationID string
	LastCommit  string
	Processed   int
	Total       int
	SourcePath  string
	TargetPath  string
	LastUpdated time.Time
	Status      string
}

// StateDB provides SQLite-based state persistence
type StateDB struct {
	db *sql.DB
}

// NewStateDB creates a new state database
func NewStateDB(path string) (*StateDB, error) {
	// Ensure parent directory exists to prevent I/O errors during rapid test execution
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: failed to close database after ping error: %v", closeErr)
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings for better reliability
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Set SQLite pragmas for better reliability during rapid test execution
	// These must be set via EXEC statements, not DSN parameters, to avoid file path issues
	pragmas := []string{
		"PRAGMA journal_mode=DELETE;", // Use DELETE mode (default) - more reliable for rapid creation
		"PRAGMA busy_timeout=5000;",   // Wait up to 5 seconds for locks
		"PRAGMA synchronous=OFF;",     // Disable sync for test reliability
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			if closeErr := db.Close(); closeErr != nil {
				log.Printf("Warning: failed to close database after pragma error: %v", closeErr)
			}
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	// Create schema - execute statements individually to avoid multi-statement issues
	schemaStatements := []string{
		`CREATE TABLE IF NOT EXISTS migration_state (
			migration_id TEXT PRIMARY KEY,
			last_commit TEXT,
			processed INTEGER,
			total INTEGER,
			source_path TEXT,
			target_path TEXT,
			last_updated TIMESTAMP,
			status TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_status ON migration_state(status)`,
		`CREATE INDEX IF NOT EXISTS idx_last_updated ON migration_state(last_updated)`,
	}

	for _, stmt := range schemaStatements {
		if _, err := db.Exec(stmt); err != nil {
			if closeErr := db.Close(); closeErr != nil {
				log.Printf("Warning: failed to close database after schema error: %v", closeErr)
			}
			return nil, fmt.Errorf("failed to execute schema statement: %w", err)
		}
	}

	// Execute a simple query to ensure database file is created and synchronized
	// This is important for tests that check for file existence immediately after creation
	if _, err := db.Exec("SELECT 1;"); err != nil {
		// This shouldn't fail, but if it does, the database has issues
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: failed to close database after verification error: %v", closeErr)
		}
		return nil, fmt.Errorf("failed to verify database: %w", err)
	}

	return &StateDB{db: db}, nil
}

// Save saves migration state
func (sdb *StateDB) Save(state *MigrationState) error {
	state.LastUpdated = time.Now()

	query := `
	INSERT OR REPLACE INTO migration_state
		(migration_id, last_commit, processed, total, source_path, target_path, last_updated, status)
	VALUES
		(?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := sdb.db.Exec(query,
		state.MigrationID,
		state.LastCommit,
		state.Processed,
		state.Total,
		state.SourcePath,
		state.TargetPath,
		state.LastUpdated,
		state.Status,
	)

	return err
}

// Load loads migration state
func (sdb *StateDB) Load(migrationID string) (*MigrationState, error) {
	query := `
	SELECT migration_id, last_commit, processed, total, source_path, target_path, last_updated, status
	FROM migration_state
	WHERE migration_id = ?
	`

	state := &MigrationState{}
	err := sdb.db.QueryRow(query, migrationID).Scan(
		&state.MigrationID,
		&state.LastCommit,
		&state.Processed,
		&state.Total,
		&state.SourcePath,
		&state.TargetPath,
		&state.LastUpdated,
		&state.Status,
	)

	if err != nil {
		return nil, err
	}

	return state, nil
}

// Complete marks a migration as completed
func (sdb *StateDB) Complete(migrationID string) error {
	query := `
	UPDATE migration_state
	SET status = 'completed', last_updated = ?
	WHERE migration_id = ?
	`

	_, err := sdb.db.Exec(query, time.Now(), migrationID)
	return err
}

// Delete deletes migration state
func (sdb *StateDB) Delete(migrationID string) error {
	_, err := sdb.db.Exec("DELETE FROM migration_state WHERE migration_id = ?", migrationID)
	return err
}

// History returns migration history
func (sdb *StateDB) History() ([]*MigrationState, error) {
	query := `
	SELECT migration_id, last_commit, processed, total, source_path, target_path, last_updated, status
	FROM migration_state
	ORDER BY last_updated DESC
	`

	rows, err := sdb.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var history []*MigrationState
	for rows.Next() {
		state := &MigrationState{}
		if err := rows.Scan(
			&state.MigrationID,
			&state.LastCommit,
			&state.Processed,
			&state.Total,
			&state.SourcePath,
			&state.TargetPath,
			&state.LastUpdated,
			&state.Status,
		); err != nil {
			return nil, err
		}
		history = append(history, state)
	}

	return history, nil
}

// Close closes the database connection
func (sdb *StateDB) Close() error {
	// Ensure all idle connections are closed before closing the main connection
	// This helps prevent resource leaks during rapid test execution
	sdb.db.SetMaxIdleConns(0)
	return sdb.db.Close()
}
