package database

import (
	"context"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
)

//// GetScenarioStatus retrieves only the status of a scenario by its ID
//func (d *Database) GetScenarioStatus(scenarioID string) (string, error) {
//	var status string
//	err := d.DB.QueryRow(
//		"SELECT status FROM scenarios WHERE id = $1",
//		scenarioID,
//	).Scan(&status)
//
//	return status, err
//}

// WriteHeartbeat record a heartbeat
func (d *Database) WriteHeartbeat(heartbeat models.Heartbeat) error {
	_, err := d.DB.Exec(
		"INSERT INTO heartbeats (scenario_id, status, frame, timestamp) VALUES ($1, $2, $3, $4)",
		heartbeat.ScenarioID,
		heartbeat.Action,
		heartbeat.Frame,
		heartbeat.TimeStamp,
	)

	return err
}

func (d *Database) FindStuckScenarios(ctx context.Context, interval time.Duration) ([]models.Scenario, error) {
	rows, err := d.DB.QueryContext(ctx, `
		SELECT s.id, s.status, s.video_source, s.created_at, s.updated_at
		FROM scenarios s
		LEFT JOIN (
			SELECT scenario_id, MAX(timestamp) as last_heartbeat
			FROM heartbeats
			GROUP BY scenario_id
		) h ON s.id = h.scenario_id
		WHERE s.status = $1 
		AND (h.last_heartbeat IS NULL OR h.last_heartbeat < $2)
	`, models.StatusActive, time.Now().Add(-interval))

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scenarios []models.Scenario
	for rows.Next() {
		var s models.Scenario
		err := rows.Scan(
			&s.ID,
			&s.Status,
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
