package database

import (
	"context"
	"log"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
)

// GetScenarioByID retrieves a scenario by its ID
func (d *Database) GetScenarioByID(scenarioID string) (models.Scenario, error) {
	var s models.Scenario
	err := d.DB.QueryRow(
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

// CreateScenario creates a new scenario record
func (d *Database) CreateScenario(ctx context.Context, scenario *models.Scenario) error {
	now := time.Now()
	scenario.CreatedAt = now
	scenario.UpdatedAt = now

	_, err := d.querier(ctx).Exec(
		"INSERT INTO scenarios (id, status, video_source, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
		scenario.ID,
		scenario.Status,
		scenario.VideoSource,
		scenario.CreatedAt,
		scenario.UpdatedAt,
	)

	return err
}

// UpdateScenarioStatus updates an existing scenario status in tx
func (d *Database) UpdateScenarioStatus(ctx context.Context, scenarioID string, status models.ScenarioStatus) error {
	log.Printf("UpdateScenarioStatus started %s %s\n", scenarioID, status)
	_, err := d.querier(ctx).Exec(
		"UPDATE scenarios SET status = $1, updated_at = NOW() WHERE id = $2",
		status,
		scenarioID,
	)

	return err
}
