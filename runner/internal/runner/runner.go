package runner

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/database"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/kafka"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/models"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/s3"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/services/detection"
	"github.com/samber/lo"
)

const (
	//maxScenarios            = 1
	retries                 = 5
	heartbeatInterval       = 5 * time.Second
	checkStopEventsInterval = 10 * time.Second
)

type Runner struct {
	db              *database.Database
	s3Client        *s3.Client
	detectionClient *detection.Client
	consumer        *kafka.Consumer
	producer        *kafka.Producer

	activeRunners map[string]context.CancelFunc
	mu            sync.Mutex
}

func New(db *database.Database, s3Client *s3.Client, detectionClient *detection.Client, consumer *kafka.Consumer, producer *kafka.Producer) *Runner {
	return &Runner{
		db:              db,
		s3Client:        s3Client,
		detectionClient: detectionClient,
		consumer:        consumer,
		producer:        producer,
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
			var cmd models.ScenarioCommand
			if err := json.Unmarshal(msg.Value, &cmd); err != nil {
				log.Printf("Invalid message format: %v", err)
				// Не подтверждаем сообщение при ошибке парсинга
				continue
			}
			log.Printf("Runner: received scenario command %v", cmd)

			var processErr error
			switch cmd.Action {
			case models.CommandStart:
				processErr = r.Start(ctx, cmd)
			case models.CommandStop:
				processErr = r.RegisterStopEvent(cmd.ScenarioID)
			default:
				log.Printf("Unknown command: %s", cmd.Action)
			}

			if processErr != nil {
				log.Printf("Error processing command: %v", processErr)
				// Не подтверждаем сообщение при ошибке обработки
				continue
			}

			// Подтверждаем сообщение только после успешной обработки
			msg.Session.MarkMessage(msg.Message, "")
		}
	}
}

func (r *Runner) Start(ctx context.Context, cmd models.ScenarioCommand) error {
	existScenario, err := r.db.GetScenario(cmd.ScenarioID)
	if err != nil {
		log.Printf("Database error: %v", err)
		return err
	}
	if existScenario != nil {
		if existScenario.Action == models.CommandStart && time.Now().Sub(existScenario.UpdatedAt) < heartbeatInterval*3 {
			log.Printf("Runner for %s already running", cmd.ScenarioID)
			return err
		}
	}

	if err := r.db.CreateScenario(&models.Scenario{
		ID:          cmd.ScenarioID,
		Action:      cmd.Action,
		VideoSource: cmd.VideoSource,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}); err != nil {
		log.Printf("Database error: %v", err)
		return err
	}
	log.Printf("Runner for %s created", cmd.ScenarioID)

	if err := r.producer.SendHeartbeat(models.Heartbeat{
		ScenarioID: cmd.ScenarioID,
		Action:     models.CommandStart,
		TimeStamp:  time.Now().UTC(),
	}); err != nil {
		log.Printf("Runner %s error sending live heartbeat: %v", cmd.ScenarioID, err)
		return err
	}

	r.mu.Lock()
	childCtx, cancel := context.WithCancel(ctx)
	r.activeRunners[cmd.ScenarioID] = cancel
	r.mu.Unlock()

	//if len(r.activeRunners) >= maxScenarios {
	//	log.Printf("Runner for %s max scenarios reached", cmd.ScenarioID)
	//	r.consumer.Close()
	//}

	go func() {
		defer func() {
			//if len(r.activeRunners) == maxScenarios {
			//	go r.consumer.StartListening(ctx)
			//}

			r.mu.Lock()
			delete(r.activeRunners, cmd.ScenarioID)
			r.mu.Unlock()

			log.Printf("Runner %s finished", cmd.ScenarioID)
		}()

		if err := r.processScenario(childCtx, cmd); err != nil {
			log.Printf("Runner %s error: %v", cmd.ScenarioID, err)
		}
	}()

	return nil
}

// processScenario скачивает кадры, отправляет их на детекцию и сохраняет в s3
func (r *Runner) processScenario(ctx context.Context, cmd models.ScenarioCommand) error {
	log.Printf("Runner %s: downloading frames from %s", cmd.ScenarioID, cmd.VideoSource)

	frames, err := r.s3Client.DownloadFilesFromURL(ctx, cmd.VideoSource)
	if err != nil {
		return err
	}

	processedFramesCount, err := r.s3Client.CountFilesInFolder(ctx, cmd.ScenarioID)
	if err != nil {
		return err
	}

	timer := time.NewTicker(heartbeatInterval)
	log.Printf("Runner %s: started processing from %d frame", cmd.ScenarioID, processedFramesCount)
	for idx := processedFramesCount; idx < len(frames); idx++ {
		if err := r.processFrameWithRetries(ctx, cmd, frames[idx], idx); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			if err := r.db.UpdateScenarioTimestamp(cmd.ScenarioID); err != nil {
				log.Printf("Runner %s error updating scenario timestamp: %v", cmd.ScenarioID, err)
			}

			if err := r.producer.SendHeartbeat(models.Heartbeat{
				ScenarioID: cmd.ScenarioID,
				Action:     models.CommandStart,
				Frame:      int64(idx),
				TimeStamp:  time.Now().UTC(),
			}); err != nil {
				log.Printf("Runner %s error sending live heartbeat: %v", cmd.ScenarioID, err)
			}
		default:
		}
	}

	if err := r.producer.SendHeartbeat(models.Heartbeat{
		ScenarioID: cmd.ScenarioID,
		Action:     models.CommandStop,
		Frame:      int64(len(frames)),
		TimeStamp:  time.Now().UTC(),
	}); err != nil {
		log.Printf("Runner %s error sending live heartbeat: %v", cmd.ScenarioID, err)
	}
	log.Printf("Runner %s: finished sending %d frames", cmd.ScenarioID, len(frames))
	return nil
}

func (r *Runner) processFrameWithRetries(ctx context.Context, cmd models.ScenarioCommand, frame []byte, idx int) error {
	success := false

	for attempt := 0; !success && attempt < retries; attempt++ {
		if success {
			break
		}

		select {
		case <-ctx.Done():
			log.Printf("Runner %s: received stop", cmd.ScenarioID)
			return nil
		default:
			detections, err := r.detectionClient.SendFrame(frame, cmd.ScenarioID)
			if err != nil {
				log.Printf("Runner %s: detection error: %v", cmd.ScenarioID, err)
				continue
			}

			if err := r.s3Client.SaveDetectionResults(ctx, cmd.ScenarioID, idx, detections); err != nil {
				log.Printf("Runner %s: save detection error: %v", cmd.ScenarioID, err)
				continue
			}

			success = true
		}
	}

	if !success {
		log.Printf("Runner %s: failed to process frame %d", cmd.ScenarioID, idx)
	}

	return nil
}

func (r *Runner) RegisterStopEvent(scenarioID string) error {
	if err := r.db.ChangeScenarioAction(scenarioID, models.CommandStop); err != nil {
		log.Printf("Runner %s error stopping scenario: %v", scenarioID, err)
		return err
	}

	return nil
}

func (r *Runner) ProcessStopEvent(ctx context.Context) {
	timer := time.NewTicker(checkStopEventsInterval)
	for {
		select {
		case <-timer.C:
			scenarios, err := r.db.GetInactiveScenarios(ctx)
			if err != nil {
				log.Printf("Error getting inactive scenario status: %v", err)
			}

			scenarioIDs := lo.Map(scenarios, func(s models.Scenario, _ int) string {
				return s.ID
			})

			for _, scenarioID := range scenarioIDs {
				if r.Stop(ctx, scenarioID) {
					if err := r.producer.SendHeartbeat(models.Heartbeat{
						ScenarioID: scenarioID,
						Action:     models.CommandStop,
						TimeStamp:  time.Now().UTC(),
					}); err != nil {
						log.Printf("Runner %s error sending stop heartbeat: %v", scenarioID, err)
					}
				}
			}

		default:
		}
	}
}

func (r *Runner) Stop(ctx context.Context, scenarioID string) bool {
	//if len(r.activeRunners) == maxScenarios {
	//	go r.consumer.StartListening(ctx)
	//}

	r.mu.Lock()
	defer r.mu.Unlock()

	if cancel, ok := r.activeRunners[scenarioID]; ok {
		cancel()
		log.Printf("Runner %s stopped", scenarioID)
		return true
	}

	return false
}
