package models

import "time"

type CommandAction string

const (
	CommandStart CommandAction = "start"
	CommandStop  CommandAction = "stop"
)

// Detection представляет структуру одного обнаруженного объекта
type Detection struct {
	Class string    `json:"class"`
	Score float64   `json:"score"`
	Box   []float64 `json:"box"` // [x1, y1, x2, y2]
}

type ScenarioCommand struct {
	ScenarioID  string        `json:"scenario_id"`
	Action      CommandAction `json:"action"`
	VideoSource string        `json:"video_source"`
}

type Heartbeat struct {
	ScenarioID string        `json:"ScenarioID"`
	Action     CommandAction `json:"Action"`
	Frame      int64         `json:"Frame"`
	TimeStamp  time.Time     `json:"TimeStamp"`
}

// Scenario Структура для сценариев
type Scenario struct {
	ID          string        `json:"id"`
	Action      CommandAction `json:"action"`
	VideoSource string        `json:"video_source"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
