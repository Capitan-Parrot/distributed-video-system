package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
	"github.com/gorilla/mux"
)

// UpdateScenarioStatusHandler обработчик для обновления статуса сценария
func (h *Handlers) UpdateScenarioStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scenarioID := vars["scenario_id"]
	action := models.CommandAction(r.URL.Query().Get("action"))
	if action == "" {
		http.Error(w, "action parameter is required (start/stop)", http.StatusBadRequest)
		return
	}

	// Проверка существования сценария и получение текущего статуса
	scenario, err := h.db.GetScenarioByID(scenarioID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Scenario not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}
	currentStatus := scenario.Status

	var newStatus models.ScenarioStatus
	ctx := context.Background()
	if action == models.CommandStart {
		switch currentStatus {
		case models.StatusInitStartup, models.StatusInStartupProcessing, models.StatusActive:
			http.Error(w, "Invalid transaction", http.StatusBadRequest)
			return
		case models.StatusInitShutdown:
			if err := h.db.InTx(ctx, func(ctx context.Context) error {
				_, err := h.db.MarkOutboxMessageProcessedByScenarioID(ctx, scenarioID)
				if err != nil {
					return err
				}

				err = h.db.UpdateScenarioStatus(ctx, scenarioID, models.StatusInitStartup)
				if err != nil {
					return err
				}

				err = h.db.AddToOutbox(ctx, scenarioID, models.CommandStart)
				if err != nil {
					return err
				}

				return nil
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			newStatus = models.StatusInitShutdown
		case models.StatusInShutdownProcessing, models.StatusInactive:
			if err := h.db.InTx(ctx, func(ctx context.Context) error {
				err = h.db.UpdateScenarioStatus(ctx, scenarioID, models.StatusInitStartup)
				if err != nil {
					return err
				}
				err = h.db.AddToOutbox(ctx, scenarioID, models.CommandStart)
				if err != nil {
					return err
				}

				return nil
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			newStatus = models.StatusInitStartup
		}
	} else if action == models.CommandStop {
		switch currentStatus {
		case models.StatusInitShutdown, models.StatusInShutdownProcessing, models.StatusInactive:
			http.Error(w, "Invalid transaction", http.StatusBadRequest)
			return
		case models.StatusInitStartup:
			if err := h.db.InTx(ctx, func(ctx context.Context) error {
				_, err := h.db.MarkOutboxMessageProcessedByScenarioID(ctx, scenarioID)
				if err != nil {
					return err
				}

				err = h.db.UpdateScenarioStatus(ctx, scenarioID, models.StatusInitShutdown)
				if err != nil {
					return err
				}

				err = h.db.AddToOutbox(ctx, scenarioID, models.CommandStop)
				if err != nil {
					return err
				}

				return nil
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			newStatus = models.StatusInitShutdown
		case models.StatusInStartupProcessing, models.StatusActive:
			if err := h.db.InTx(ctx, func(ctx context.Context) error {
				err = h.db.UpdateScenarioStatus(ctx, scenarioID, models.StatusInitShutdown)
				if err != nil {
					return err
				}
				return h.db.AddToOutbox(ctx, scenarioID, models.CommandStop)
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			newStatus = models.StatusInitShutdown
		}
	}
	log.Println(action, newStatus)

	// Возвращаем обновленный статус
	response := map[string]string{"id": scenarioID, "status": string(newStatus)}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
