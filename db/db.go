package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	generated "sandstorm-tracker/db/generated"

	_ "modernc.org/sqlite"
)

// DatabaseService provides database operations
type DatabaseService struct {
	db      *sql.DB
	queries *generated.Queries
}

// NewDatabaseService creates a new database service
func NewDatabaseService(dbPath string) (*DatabaseService, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := configureDatabase(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to configure database: %w", err)
	}

	if err := initializeSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &DatabaseService{
		db:      db,
		queries: generated.New(db),
	}, nil
}

func (ds *DatabaseService) GetQueries() *generated.Queries { return ds.queries }
func (ds *DatabaseService) GetDB() *sql.DB                 { return ds.db }
func (ds *DatabaseService) Close() error                   { return ds.db.Close() }

func configureDatabase(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = 1000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("failed to set pragma %s: %w", p, err)
		}
	}
	return nil
}

func initializeSchema(db *sql.DB) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_version'`).Scan(&count); err != nil {
		return fmt.Errorf("failed to check schema_version: %w", err)
	}
	if count == 0 {
		return createFullSchema(db)
	}
	return nil
}

func createFullSchema(db *sql.DB) error {
	schema := `
-- Simplified servers table
CREATE TABLE IF NOT EXISTS servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    external_id TEXT UNIQUE,
    name TEXT,
    path TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Server logs
CREATE TABLE IF NOT EXISTS server_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER NOT NULL,
    open_time DATETIME NOT NULL,
    log_path TEXT NOT NULL,
    lines_processed INTEGER DEFAULT 0,
    file_size_bytes INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
);
-- Players
CREATE TABLE IF NOT EXISTS players (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    external_id TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Matches
CREATE TABLE IF NOT EXISTS matches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER NOT NULL,
    map TEXT,
    mode TEXT,
    winner_team INTEGER,
    start_time DATETIME,
    end_time DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
);

-- Match participant
CREATE TABLE IF NOT EXISTS match_participant (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id INTEGER NOT NULL,
    match_id INTEGER NOT NULL,
    join_time DATETIME,
    leave_time DATETIME,
    team INTEGER,
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE
);

-- Kills
CREATE TABLE IF NOT EXISTS kills (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER NOT NULL,
    match_id INTEGER,
    weapon_name TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    killer_id INTEGER,
    victim_name TEXT,
    is_team_kill BOOLEAN DEFAULT 0,
    is_suicide BOOLEAN DEFAULT 0,
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE SET NULL,
    FOREIGN KEY (killer_id) REFERENCES players(id) ON DELETE SET NULL
);

-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_version (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version INTEGER NOT NULL,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO schema_version (version) VALUES (1);
`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	return nil
}
