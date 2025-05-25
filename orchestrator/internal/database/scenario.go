package database

import (
	"database/sql"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
)

// GetScenarioByID retrieves a scenario by its ID
func (d *Database) GetScenarioByID(scenarioID string) (models.Scenario, error) {
	var s models.Scenario
	err := d.Db.QueryRow(
		"SELECT id, status, video_source, created_at, updated_at FROM scenarios WHERE id = $1",
		scenarioID,
	).Scan(
		&s.ID,
		&s.Status,
		&s.VideoSource,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		return models.Scenario{}, err
	}

	return s, nil
}

// GetScenarioStatus retrieves only the status of a scenario by its ID
func (d *Database) GetScenarioStatus(scenarioID string) (string, error) {
	var status string
	err := d.Db.QueryRow(
		"SELECT status FROM scenarios WHERE id = $1",
		scenarioID,
	).Scan(&status)

	return status, err
}

// CreateScenario creates a new scenario record
func (d *Database) CreateScenario(tx *sql.Tx, scenario *models.Scenario) error {
	now := time.Now()
	scenario.CreatedAt = now
	scenario.UpdatedAt = now

	_, err := tx.Exec(
		"INSERT INTO scenarios (id, status, video_source, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
		scenario.ID,
		scenario.Status,
		scenario.VideoSource,
		scenario.CreatedAt,
		scenario.UpdatedAt,
	)

	return err
}

// UpdateScenario updates an existing scenario
func (d *Database) UpdateScenario(scenario *models.Scenario) error {
	scenario.UpdatedAt = time.Now()

	_, err := d.Db.Exec(
		"UPDATE scenarios SET status = $1, video_source = $2, updated_at = $3 WHERE id = $4",
		scenario.Status,
		scenario.VideoSource,
		scenario.UpdatedAt,
		scenario.ID,
	)

	return err
}
