package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
	"github.com/google/uuid"
)

func (h *Handlers) CreateScenarioHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "Could not parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Video file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Получаем размер файла
	size := header.Size
	if size == 0 {
		http.Error(w, "Video file is empty", http.StatusBadRequest)
		return
	}

	// Генерация ID и имени объекта
	id := uuid.New().String()
	objectName := fmt.Sprintf("%s.mp4", id)

	// Загружаем напрямую в S3
	videoURL, err := h.s3.UploadFileStream("videos", objectName, file, size)
	if err != nil {
		http.Error(w, "Failed to upload to S3", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	initialStatus := models.StatusInitStartup

	tx, err := h.db.BeginTransaction()
	if err != nil {
		http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
		return
	}

	scenario := models.Scenario{
		ID:          id,
		Status:      initialStatus,
		VideoSource: videoURL,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.db.CreateScenario(tx, &scenario); err != nil {
		tx.Rollback()
		http.Error(w, "Failed to insert scenario", http.StatusInternalServerError)
		return
	}

	if err := h.db.AddToOutbox(tx, id, models.CommandStart); err != nil {
		tx.Rollback()
		http.Error(w, "Failed to add to outbox", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":     id,
		"status": initialStatus,
	})
}
