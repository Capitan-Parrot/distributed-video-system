package outbox

import (
	"context"
	"log"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/database"
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/kafka"
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
				err = producer.SendOutboxMessageToKafka(&msg)
				if err != nil {
					log.Printf("Failed to send message to Kafka: %v", err)
					continue
				}

				// Отмечаем сообщение как обработанное
				if err := db.MarkOutboxMessageAsProcessed(msg.ID); err != nil {
					log.Printf("Failed to mark outbox message as processed: %v", err)
				}
			}
		}
	}
}
