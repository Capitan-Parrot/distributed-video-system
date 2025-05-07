package models

import (
	"encoding/json"
	"time"
)

// Scenario Структура для сценариев
type Scenario struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	VideoSource string    `json:"video_source"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Prediction Структура для предсказаний
type Prediction struct {
	ID         string          `json:"id"`
	ScenarioID string          `json:"scenario_id"`
	Timestamp  time.Time       `json:"timestamp"`
	Data       json.RawMessage `json:"data"`
}

// ScenarioCreate Структура для создания сценария
type ScenarioCreate struct {
	VideoSource string          `json:"video_source"`
	Config      json.RawMessage `json:"config,omitempty"`
}

// StatusUpdate Структура для обновления статуса
type StatusUpdate struct {
	Status string `json:"status"`
}

// OutboxMessage Структура для транзакционного outbox
type OutboxMessage struct {
	ID          string     `json:"id"`
	ScenarioID  string     `json:"scenario_id"`
	Action      string     `json:"action"`
	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at"`
	VideoSource string     `json:"video_source"`
}

// Константы статусов
const (
	StatusInitStartup          = "init_startup"
	StatusInStartupProcessing  = "in_startup_processing"
	StatusActive               = "active"
	StatusInitShutdown         = "init_shutdown"
	StatusInShutdownProcessing = "in_shutdown_processing"
	StatusInactive             = "inactive"
)

type CommandAction string

const (
	CommandStart CommandAction = "start"
	CommandStop  CommandAction = "stop"
)
