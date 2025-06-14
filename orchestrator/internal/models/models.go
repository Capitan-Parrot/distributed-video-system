package models

import (
	"encoding/json"
	"time"
)

// ScenarioStatus Константы статусов
type ScenarioStatus string

const (
	StatusInitStartup          ScenarioStatus = "init_startup"
	StatusInStartupProcessing  ScenarioStatus = "in_startup_processing"
	StatusActive               ScenarioStatus = "active"
	StatusInitShutdown         ScenarioStatus = "init_shutdown"
	StatusInShutdownProcessing ScenarioStatus = "in_shutdown_processing"
	StatusInactive             ScenarioStatus = "inactive"
)

// Scenario Структура для сценариев
type Scenario struct {
	ID          string         `json:"id"`
	Status      ScenarioStatus `json:"status"`
	VideoSource string         `json:"video_source"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
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
	ID          string        `json:"id"`
	ScenarioID  string        `json:"scenario_id"`
	Action      CommandAction `json:"action"`
	CreatedAt   time.Time     `json:"created_at"`
	ProcessedAt *time.Time    `json:"processed_at"`
	VideoSource string        `json:"video_source"`
}

type CommandAction string

const (
	CommandStart CommandAction = "start"
	CommandStop  CommandAction = "stop"
)

type Heartbeat struct {
	ScenarioID string        `json:"ScenarioID"`
	Action     CommandAction `json:"Action"`
	Frame      int64         `json:"Frame"`
	TimeStamp  time.Time     `json:"TimeStamp"`
}
