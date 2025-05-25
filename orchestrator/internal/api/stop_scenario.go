package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
	"github.com/gorilla/mux"
)

// UpdateScenarioStatusHandler обработчик для обновления статуса сценария
func (h *Handlers) UpdateScenarioStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scenarioID := vars["scenario_id"]

	// Проверка существования сценария и получение текущего статуса
	currentStatus, err := h.db.GetScenarioStatus(scenarioID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Scenario not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}
	var newStatus models.CommandAction
	if models.CommandAction(currentStatus) == models.CommandStart {
		newStatus = models.CommandStart
	} else {
		newStatus = models.CommandStop
	}

	// Начинаем транзакцию
	now := time.Now()
	tx, err := h.db.BeginTransaction()
	if err != nil {
		http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
		return
	}

	// Обновление статуса в БД
	_, err = tx.Exec(
		"UPDATE scenarios SET status = $1, updated_at = $2 WHERE id = $3", newStatus, now, scenarioID,
	)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update scenario status", http.StatusInternalServerError)
		return
	}

	err = h.db.AddToOutbox(tx, scenarioID, models.CommandStop)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to add to outbox", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Возвращаем обновленный статус
	response := map[string]string{"id": scenarioID, "status": string(newStatus)}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		// TODO logger
	}
}
