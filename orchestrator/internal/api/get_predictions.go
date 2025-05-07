package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
	"github.com/gorilla/mux"
)

// GetPredictionsHandler обработчик для получения предсказаний по ID сценария
func (h *Handlers) GetPredictionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scenarioID := vars["scenario_id"]

	// Проверка существования сценария
	var exists bool
	err := h.db.Db.QueryRow("SELECT 1 FROM scenarios WHERE id = ?", scenarioID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Scenario not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Получение всех предсказаний для сценария
	rows, err := h.db.Db.Query(
		"SELECT id, scenario_id, timestamp, data FROM predictions WHERE scenario_id = ? ORDER BY timestamp DESC LIMIT 100",
		scenarioID,
	)
	if err != nil {
		http.Error(w, "Failed to fetch predictions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var predictions []models.Prediction
	for rows.Next() {
		var prediction models.Prediction
		if err := rows.Scan(&prediction.ID, &prediction.ScenarioID, &prediction.Timestamp, &prediction.Data); err != nil {
			http.Error(w, "Failed to parse predictions", http.StatusInternalServerError)
			return
		}
		predictions = append(predictions, prediction)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(predictions)
}
