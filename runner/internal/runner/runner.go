package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/kafka"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/models"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/s3"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/services/detection"
)

const (
	frameSkip    = 10
	tmpVideoPath = "/tmp"
)

type ScenarioCommand struct {
	ScenarioID  string               `json:"scenario_id"`
	Action      models.CommandAction `json:"action"`
	VideoSource string               `json:"video_source"`
}

type Runner struct {
	s3Client        *s3.Client
	detectionClient *detection.Client
	consumer        *kafka.Consumer

	activeRunners map[string]context.CancelFunc
	mu            sync.Mutex
}

func New(s3Client *s3.Client, detectionClient *detection.Client, consumer *kafka.Consumer) *Runner {
	return &Runner{
		s3Client:        s3Client,
		detectionClient: detectionClient,
		consumer:        consumer,
		activeRunners:   make(map[string]context.CancelFunc),
	}
}

func (r *Runner) ListenAndRun(ctx context.Context) {
	log.Println("Runner: listening for Kafka commands")
	for {
		select {
		case <-ctx.Done():
			log.Println("Runner: shutting down")
			return
		case msg := <-r.consumer.Messages():
			var cmd ScenarioCommand
			if err := json.Unmarshal(msg, &cmd); err != nil {
				log.Printf("Invalid message format: %v", err)
				continue
			}

			switch cmd.Action {
			case models.CommandStart:
				r.Start(ctx, cmd)
			case models.CommandStop:
				r.Stop(cmd.ScenarioID)
			default:
				log.Printf("Unknown command: %s", cmd.Action)
			}
		}
	}
}

func (r *Runner) Start(ctx context.Context, cmd ScenarioCommand) {
	r.mu.Lock()
	if _, exists := r.activeRunners[cmd.ScenarioID]; exists {
		r.mu.Unlock()
		log.Printf("Runner for %s already running", cmd.ScenarioID)
		return
	}

	childCtx, cancel := context.WithCancel(ctx)
	r.activeRunners[cmd.ScenarioID] = cancel
	r.mu.Unlock()

	go func() {
		defer func() {
			r.mu.Lock()
			delete(r.activeRunners, cmd.ScenarioID)
			r.mu.Unlock()
			log.Printf("Runner %s finished", cmd.ScenarioID)
		}()

		if err := r.processScenario(childCtx, cmd); err != nil {
			log.Printf("Runner %s error: %v", cmd.ScenarioID, err)
		}
	}()
}

func (r *Runner) Stop(scenarioID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cancel, ok := r.activeRunners[scenarioID]; ok {
		cancel()
		log.Printf("Runner %s stopped", scenarioID)
	} else {
		log.Printf("Runner %s not found", scenarioID)
	}
}

// processScenario скачивает видео, извлекает кадры и отправляет их на детекцию
func (r *Runner) processScenario(ctx context.Context, cmd ScenarioCommand) error {
	log.Printf("Runner %s: downloading video from %s", cmd.ScenarioID, cmd.VideoSource)

	// Скачиваем видео с S3
	videoData, err := r.s3Client.DownloadFileFromURL(cmd.VideoSource)
	if err != nil {
		return fmt.Errorf("download error: %w", err)
	}

	// Создаём временный файл для видео
	tmpPath := filepath.Join(tmpVideoPath, "video_"+cmd.ScenarioID+".mp4")
	if err := os.WriteFile(tmpPath, videoData, 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	defer os.Remove(tmpPath)

	// Используем ffmpeg для извлечения кадров
	return r.extractFramesFromVideo(ctx, tmpPath, cmd)
}

// extractFramesFromVideo извлекает кадры с использованием ffmpeg
func (r *Runner) extractFramesFromVideo(ctx context.Context, videoPath string, cmd ScenarioCommand) error {
	// Запускаем ffmpeg процесс для извлечения кадров
	cmdLine := exec.Command(
		"ffmpeg", "-i", videoPath, "-vf", "fps=1", // Извлекаем 1 кадр в секунду
		"-q:v", "2", "-f", "image2pipe", "pipe:1",
	)

	// Запускаем процесс
	stdoutPipe, err := cmdLine.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe error: %w", err)
	}

	if err := cmdLine.Start(); err != nil {
		return fmt.Errorf("ffmpeg start error: %w", err)
	}
	defer cmdLine.Wait()

	frameCount := 0
	buf := make([]byte, 1024*1024) // Буфер для чтения данных

	for {
		select {
		case <-ctx.Done():
			log.Printf("Runner %s: received stop", cmd.ScenarioID)
			return nil
		default:
			n, err := stdoutPipe.Read(buf)
			if err != nil && err.Error() != "EOF" {
				log.Printf("Runner %s: error reading frame data: %v", cmd.ScenarioID, err)
				continue
			}
			if n == 0 {
				log.Printf("Runner %s: video extraction finished", cmd.ScenarioID)
				return nil
			}

			// Пропускаем кадры
			frameCount++
			if frameCount%frameSkip != 0 {
				continue
			}

			// Отправляем кадр в детекцию
			frameData := buf[:n]
			go func(frame []byte) {
				if err := r.detectionClient.SendFrame(frame, cmd.ScenarioID); err != nil {
					log.Printf("Runner %s: detection error: %v", cmd.ScenarioID, err)
				}
			}(frameData)
		}
	}
}
