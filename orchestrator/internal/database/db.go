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
		status TEXT NOT NULL,
		video_source TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS heartbeats (
		id SERIAL PRIMARY KEY,
		scenario_id TEXT NOT NULL,
		status TEXT NOT NULL,
		frame INTEGER NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS outbox (
		id TEXT PRIMARY KEY,
		scenario_id TEXT NOT NULL,
		action TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		processed_at TIMESTAMP,
		FOREIGN KEY (scenario_id) REFERENCES scenarios(id)
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
