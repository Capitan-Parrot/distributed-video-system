package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/utils"
	"github.com/gorilla/mux"
)

// UpdateScenarioStatusHandler обработчик для обновления статуса сценария
func (h *Handlers) UpdateScenarioStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scenarioID := vars["scenario_id"]

	var statusUpdate models.StatusUpdate
	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

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

	// Проверка допустимости перехода статуса
	if !utils.IsValidStatusTransition(currentStatus, statusUpdate.Status) {
		http.Error(w, fmt.Sprintf("Invalid status transition from %s to %s", currentStatus, statusUpdate.Status), http.StatusBadRequest)
		return
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
		"UPDATE scenarios SET status = ?, updated_at = ? WHERE id = ?",
		statusUpdate.Status, now, scenarioID,
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
	response := map[string]string{"id": scenarioID, "status": statusUpdate.Status}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		// TODO logger
	}
}
