package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/models"
)

func (d *Database) CreateScenario(scenario *models.Scenario) error {
	now := time.Now()
	scenario.CreatedAt = now
	scenario.UpdatedAt = now

	_, err := d.DB.Exec(
		`INSERT INTO scenarios (id, action, video_source, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) 
			 	ON CONFLICT (id) DO UPDATE SET action = $6, updated_at = NOW()`,
		scenario.ID,
		scenario.Action,
		scenario.VideoSource,
		scenario.CreatedAt,
		scenario.UpdatedAt,
		models.CommandStart,
	)

	return err
}

func (d *Database) GetScenario(scenarioID string) (*models.Scenario, error) {
	row := d.DB.QueryRow(`
		SELECT id, action, video_source, created_at, updated_at
		FROM scenarios
		WHERE id = $1
	`, scenarioID)

	var scenario models.Scenario
	err := row.Scan(
		&scenario.ID,
		&scenario.Action,
		&scenario.VideoSource,
		&scenario.CreatedAt,
		&scenario.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Сценарий не найден - это не ошибка
		}
		return nil, fmt.Errorf("failed to get active scenario: %w", err)
	}

	return &scenario, nil
}

// GetInactiveScenarios retrieves stopped scenarios
func (d *Database) GetInactiveScenarios(ctx context.Context) ([]models.Scenario, error) {
	rows, err := d.DB.QueryContext(ctx, `
		SELECT id, action, video_source, created_at, updated_at
		FROM scenarios
		WHERE action = $1
	`, models.CommandStop)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scenarios []models.Scenario
	for rows.Next() {
		var s models.Scenario
		err := rows.Scan(
			&s.ID,
			&s.Action,
			&s.VideoSource,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, s)
	}

	return scenarios, nil
}

func (d *Database) ChangeScenarioAction(scenarioID string, newAction models.CommandAction) error {
	now := time.Now()

	_, err := d.DB.Exec(
		"UPDATE scenarios SET action = $1, updated_at = $2 WHERE id = $3",
		newAction,
		now,
		scenarioID,
	)

	return err
}

func (d *Database) UpdateScenarioTimestamp(scenarioID string) error {
	now := time.Now()

	_, err := d.DB.Exec(
		"UPDATE scenarios SET updated_at = $1 WHERE id = $2",
		now,
		scenarioID,
	)

	return err
}
