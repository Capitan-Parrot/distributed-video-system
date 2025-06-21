package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

	size := header.Size
	if size == 0 {
		http.Error(w, "Video file is empty", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	tempDir := os.TempDir()

	// Сохраняем видео во временный файл
	videoPath := filepath.Join(tempDir, fmt.Sprintf("%s.mp4", id))
	tempFile, err := os.Create(videoPath)
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(videoPath)
	defer tempFile.Close()
	if _, err := io.Copy(tempFile, file); err != nil {
		http.Error(w, "Failed to save video file", http.StatusInternalServerError)
		return
	}
	tempFile.Close()

	// Извлекаем кадры из видео
	framesPath := filepath.Join(tempDir, fmt.Sprintf("frames_%s", id))
	if err := os.MkdirAll(framesPath, 0755); err != nil {
		http.Error(w, fmt.Sprintf("failed to create frames directory: %v", err), http.StatusInternalServerError)
	}
	defer os.RemoveAll(framesPath)
	frames, err := extractFrames(framesPath, videoPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to extract frames: %v", err), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	// Сохраняем в S3
	err = h.saveFrames(ctx, id, frames)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save frames: %v", err), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	initialStatus := models.StatusInitStartup

	scenario := models.Scenario{
		ID:          id,
		Status:      initialStatus,
		VideoSource: fmt.Sprintf("frames/%s", id),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.db.InTx(ctx, func(ctx context.Context) error {
		if err := h.db.CreateScenario(ctx, &scenario); err != nil {
			return fmt.Errorf("failed to insert scenario: %w", err)
		}

		if err := h.db.AddToOutbox(ctx, scenario.ID, models.CommandStart); err != nil {
			return fmt.Errorf("failed to add to outbox: %w", err)
		}

		return nil
	}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create scenario: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     id,
		"status": initialStatus,
	})
}

// extractFrames извлекает кадры из видео и загружает их в S3
func extractFrames(framesPath, videoPath string) ([]string, error) {
	t := time.Now()
	fmt.Println("extractFrames started")
	// Извлекаем кадры с помощью ffmpeg
	framePattern := filepath.Join(framesPath, "frame_%04d.jpg")
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vf", "fps=3",
		"-q:v", "2", // Качество JPEG
		framePattern,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg failed: %w, stderr: %s", err, stderr.String())
	}
	log.Println(time.Now().Sub(t))

	files, err := filepath.Glob(filepath.Join(framesPath, "frame_*.jpg"))
	if err != nil {
		return nil, fmt.Errorf("failed to list frame files: %w", err)
	}

	return files, nil
}

func (h *Handlers) saveFrames(ctx context.Context, scenarioID string, files []string) error {
	frameCount := len(files)
	if frameCount == 0 {
		return fmt.Errorf("no frames extracted from video")
	}

	// Загружаем каждый кадр в S3
	for _, framePath := range files {
		frameFile, err := os.Open(framePath)
		if err != nil {
			return fmt.Errorf("failed to open frame file %s: %w", framePath, err)
		}

		frameInfo, err := frameFile.Stat()
		if err != nil {
			frameFile.Close()
			return fmt.Errorf("failed to get frame file info: %w", err)
		}

		// Имя файла в S3: frames/{scenarioID}/frame_0001.jpg
		fileName := filepath.Base(framePath)
		objectName := fmt.Sprintf("%s/%s", scenarioID, fileName)

		_, err = h.s3.UploadFileStream(ctx, "frames", objectName, frameFile, frameInfo.Size())
		frameFile.Close()

		if err != nil {
			return fmt.Errorf("failed to upload frame %s to S3: %w", fileName, err)
		}
	}

	return nil
}
