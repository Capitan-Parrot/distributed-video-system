package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// GetScenarioStatusHandler обработчик для получения информации о статусе сценария
func (h *Handlers) GetScenarioStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scenarioID := vars["scenario_id"]

	row, err := h.db.GetScenarioByID(scenarioID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(row)
}
