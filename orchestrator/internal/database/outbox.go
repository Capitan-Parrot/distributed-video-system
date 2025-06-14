package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
	"github.com/google/uuid"
)

// AddToOutbox adds a message to the transactional outbox
func (d *Database) AddToOutbox(ctx context.Context, scenarioID string, action models.CommandAction) error {
	_, err := d.querier(ctx).Exec(
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
	rows, err := d.DB.Query(`
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
			&m.VideoSource,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, nil
}

func (d *Database) MarkOutboxMessageProcessedByScenarioID(ctx context.Context, scenarioID string) (bool, error) {
	// Используем ExecContext для выполнения UPDATE
	result, err := d.querier(ctx).ExecContext(ctx, `
        UPDATE outbox 
        SET processed_at = NOW()
        WHERE id = (
            SELECT id 
            FROM outbox 
            WHERE processed_at IS NULL 
              AND scenario_id = $1
            ORDER BY created_at
            LIMIT 1
            FOR UPDATE SKIP LOCKED  -- Блокируем только выбранную строку
        )
        RETURNING id
    `, scenarioID)
	if err != nil {
		return false, fmt.Errorf("failed to mark outbox message as processed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

// MarkOutboxMessageAsProcessed marks an outbox message as processed
func (d *Database) MarkOutboxMessageAsProcessed(id string) error {
	_, err := d.DB.Exec(
		"UPDATE outbox SET processed_at = $1 WHERE id = $2",
		time.Now(),
		id,
	)
	return err
}
