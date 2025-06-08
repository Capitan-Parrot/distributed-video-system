package models

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
