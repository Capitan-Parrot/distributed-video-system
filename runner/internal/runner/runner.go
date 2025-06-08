package runner

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/kafka"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/models"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/s3"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/services/detection"
)

const (
	retries = 5
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

// processScenario скачивает кадры, отправляет их на детекцию и сохраняет в s3
func (r *Runner) processScenario(ctx context.Context, cmd ScenarioCommand) error {
	log.Printf("Runner %s: downloading frames from %s", cmd.ScenarioID, cmd.VideoSource)

	frames, err := r.s3Client.DownloadFilesFromURL(cmd.VideoSource)
	if err != nil {
		return err
	}

	processedFramesCount, err := r.s3Client.CountFilesInFolder(cmd.ScenarioID)
	if err != nil {
		return err
	}

	for idx := processedFramesCount; idx < len(frames); idx++ {
		success := false

		for attempt := 0; !success && attempt < retries; attempt++ {
			select {
			case <-ctx.Done():
				log.Printf("Runner %s: received stop", cmd.ScenarioID)
				return nil
			default:
				detections, err := r.detectionClient.SendFrame(frames[idx], cmd.ScenarioID)
				if err != nil {
					log.Printf("Runner %s: detection error: %v", cmd.ScenarioID, err)
					continue
				}

				if err := r.s3Client.SaveDetectionResults(cmd.ScenarioID, idx, detections); err != nil {
					log.Printf("Runner %s: save detection error: %v", cmd.ScenarioID, err)
					continue
				}

				success = true
			}
		}

		if !success {
			log.Printf("Runner %s: failed to process frame %d", cmd.ScenarioID, idx)
		}
	}

	log.Printf("Runner %s: finished sending %d frames", cmd.ScenarioID, len(frames))
	return nil
}
