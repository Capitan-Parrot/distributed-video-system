package database

import (
	"database/sql"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
	"github.com/google/uuid"
)

// AddToOutbox adds a message to the transactional outbox
func (d *Database) AddToOutbox(tx *sql.Tx, scenarioID string, action models.CommandAction) error {
	_, err := tx.Exec(
		"INSERT INTO outbox (id, scenario_id, action, created_at) VALUES ($1, $2, $3, $4)",
		uuid.New().String(),
		scenarioID,
		action,
		time.Now(),
	)

	return err
}

// GetPendingOutboxMessages retrieves unprocessed outbox messages
func (d *Database) GetPendingOutboxMessages(limit int) ([]models.OutboxMessage, error) {
	rows, err := d.Db.Query(`
		SELECT 
			o.id, o.scenario_id, o.action, o.created_at,
			s.video_source
		FROM outbox o
		JOIN scenarios s ON o.scenario_id = s.id
		WHERE o.processed_at IS NULL
		ORDER BY o.created_at
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.OutboxMessage
	for rows.Next() {
		var m models.OutboxMessage
		err := rows.Scan(
			&m.ID,
			&m.ScenarioID,
			&m.Action,
			&m.CreatedAt,
			&m.VideoSource, // добавляем поле
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, nil
}

// MarkOutboxMessageAsProcessed marks an outbox message as processed
func (d *Database) MarkOutboxMessageAsProcessed(id string) error {
	_, err := d.Db.Exec(
		"UPDATE outbox SET processed_at = $1 WHERE id = $2",
		time.Now(),
		id,
	)
	return err
}
