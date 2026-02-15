// Package storage provides state persistence for migrations.
package storage

import (
	"database/sql"
	"fmt"
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
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings for better reliability
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Create schema
	schema := `
	CREATE TABLE IF NOT EXISTS migration_state (
		migration_id TEXT PRIMARY KEY,
		last_commit TEXT,
		processed INTEGER,
		total INTEGER,
		source_path TEXT,
		target_path TEXT,
		last_updated TIMESTAMP,
		status TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_status ON migration_state(status);
	CREATE INDEX IF NOT EXISTS idx_last_updated ON migration_state(last_updated);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
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
	defer rows.Close()

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
	return sdb.db.Close()
}
