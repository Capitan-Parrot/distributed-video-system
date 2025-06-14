package database

import (
	"database/sql"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Database represents the database connection and operations
type Database struct {
	DB *sql.DB
}

// New creates a new Database instance
func New(dsn string) (*Database, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}

// Init creates the required tables if they don't exist
func (d *Database) Init() error {
	createTables := `
	CREATE TABLE IF NOT EXISTS scenarios (
		id TEXT PRIMARY KEY,
		action TEXT NOT NULL,
		video_source TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL
	);
	`

	_, err := d.DB.Exec(createTables)
	return err
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.DB.Close()
}

// BeginTransaction starts a new database transaction
func (d *Database) BeginTransaction() (*sql.Tx, error) {
	return d.DB.Begin()
}
