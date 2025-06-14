package outbox

import (
	"context"
	"log"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/database"
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/kafka"
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
)

func StartOutboxDispatcher(ctx context.Context, db *database.Database, brokers []string, topic string, interval time.Duration) {
	producer, err := kafka.NewKafkaProducer(brokers, topic)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Producer.Close()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Outbox dispatcher stopped")
			return
		case <-ticker.C:
			// Читаем непрочитанные сообщения
			messages, err := db.GetPendingOutboxMessages(1)
			if err != nil {
				log.Printf("Error fetching outbox messages: %v", err)
				continue
			}

			for _, msg := range messages {
				// Отправляем сообщение в Kafka
				err = producer.SendOutboxMessageToKafka(&msg)
				if err != nil {
					log.Printf("Failed to send message to Kafka: %v", err)
					continue
				}

				// Начинаем транзакцию
				if err := db.InTx(ctx, func(ctx context.Context) error {
					// Отмечаем сообщение как обработанное
					if err := db.MarkOutboxMessageAsProcessed(msg.ID); err != nil {
						log.Printf("Failed to mark outbox message as processed: %v", err)
					}

					// Обновляем статус
					var status models.ScenarioStatus
					if msg.Action == models.CommandStart {
						status = models.StatusInStartupProcessing
					} else {
						status = models.StatusInShutdownProcessing
					}
					if err := db.UpdateScenarioStatus(ctx, msg.ScenarioID, status); err != nil {
						log.Printf("Failed to update scenario status: %v", err)
					}

					return nil
				}); err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}
