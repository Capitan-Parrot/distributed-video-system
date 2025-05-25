package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

// GetScenarioStatusHandler обработчик для получения информации о статусе сценария
func (h *Handlers) GetScenarioStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scenarioID := vars["scenario_id"]

	row, err := h.db.GetScenarioByID(scenarioID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Scenario not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(row)
}
