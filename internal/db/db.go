package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
    "embed"
	generated "sandstorm-tracker/internal/db/generated"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var embeddedSchema embed.FS

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
    schemaBytes, err := embeddedSchema.ReadFile("schema.sql")
    if err != nil {
        return fmt.Errorf("failed to read embedded schema: %w", err)
    }
    schema := string(schemaBytes)
	if _, err := db.Exec(schema); err != nil {
        return fmt.Errorf("failed to create schema: %w", err)
    }
    return nil
}
