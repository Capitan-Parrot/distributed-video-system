package kafka

import (
	"context"
	"log"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
	"github.com/IBM/sarama"
	"github.com/goccy/go-json"
)

type db interface {
	GetScenarioByID(scenarioID string) (models.Scenario, error)
	WriteHeartbeat(models.Heartbeat) error
	UpdateScenarioStatus(ctx context.Context, scenarioID string, status models.ScenarioStatus) error
}

// Consumer оборачивает Sarama ConsumerGroup
type Consumer struct {
	db
	group    sarama.ConsumerGroup
	topic    string
	messages chan []byte
	closed   chan struct{}
}

// NewConsumer создаёт и возвращает новый Consumer
func NewConsumer(brokers []string, groupID, topic string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		group:    group,
		topic:    topic,
		messages: make(chan []byte),
		closed:   make(chan struct{}),
	}, nil
}

func (c *Consumer) StartListening(ctx context.Context, db db) {
	handler := &consumerGroupHandler{
		db:     db,
		closed: c.closed,
	}

	go func() {
		retryDelay := time.Second * 5
		for {
			select {
			case <-ctx.Done():
				log.Println("Consumer: context cancelled, stopping")
				return
			case <-c.closed:
				log.Println("Consumer: received close signal, stopping")
				return
			default:
				log.Println("Consumer: starting consumption cycle")
				err := c.group.Consume(ctx, []string{c.topic}, handler)
				if err != nil {
					log.Printf("Consume error: %v, retrying in %v", err, retryDelay)
					select {
					case <-ctx.Done():
						return
					case <-c.closed:
						return
					case <-time.After(retryDelay):
					}
					continue
				}

				if ctx.Err() != nil {
					return
				}
			}
		}
	}()
}

// Close останавливает потребитель и освобождает ресурсы
func (c *Consumer) Close() error {
	close(c.closed)
	return c.group.Close()
}

// Messages возвращает канал для чтения сообщений
func (c *Consumer) Messages() <-chan []byte {
	return c.messages
}

type consumerGroupHandler struct {
	db     db
	closed <-chan struct{}
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			var heartbeat models.Heartbeat
			if err := json.Unmarshal(msg.Value, &heartbeat); err != nil {
				log.Printf("Invalid message format: %v", err)
			}

			ctx := context.Background()

			scenario, err := h.db.GetScenarioByID(heartbeat.ScenarioID)
			if err != nil {
				log.Printf("Error getting scenario: %v", err)
				return err
			}

			if heartbeat.Action == models.CommandStart && scenario.Status == models.StatusInStartupProcessing {
				log.Printf("Starting heartbeat for scenario %v", scenario.ID)
				if err := h.db.UpdateScenarioStatus(ctx, heartbeat.ScenarioID, models.StatusActive); err != nil {
					log.Printf("Failed to update scenario status in DB: %v", err)
					continue
				}
			} else if heartbeat.Action == models.CommandStop &&
				(scenario.Status == models.StatusInShutdownProcessing || scenario.Status == models.StatusActive) {
				if err := h.db.UpdateScenarioStatus(ctx, heartbeat.ScenarioID, models.StatusInactive); err != nil {
					log.Printf("Failed to update scenario status in DB: %v", err)
					continue
				}
			}

			if err := h.db.WriteHeartbeat(heartbeat); err != nil {
				log.Printf("Failed to write message to DB: %v", err)
				continue
			}

			// Подтверждаем обработку сообщения
			sess.MarkMessage(msg, "")
		case <-sess.Context().Done():
			return nil
		case <-h.closed:
			return nil
		}
	}
}
