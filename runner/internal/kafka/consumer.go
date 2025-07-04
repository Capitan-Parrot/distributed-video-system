package kafka

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// Consumer оборачивает Sarama ConsumerGroup
type Consumer struct {
	group         sarama.ConsumerGroup
	topic         string
	messages      chan consumerMessage
	closed        chan struct{}
	cancelConsume context.CancelFunc // для остановки потребления
	consumeMutex  sync.Mutex         // для безопасного управления потреблением
	running       bool               // флаг, указывающий, работает ли консьюмер
}

// consumerMessage содержит сообщение и сессию для подтверждения
type consumerMessage struct {
	Value   []byte
	Session sarama.ConsumerGroupSession
	Message *sarama.ConsumerMessage
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
		messages: make(chan consumerMessage),
		closed:   make(chan struct{}),
	}, nil
}

// Pause временно отключает консьюмер от группы
func (c *Consumer) Pause() {
	c.consumeMutex.Lock()
	defer c.consumeMutex.Unlock()

	if c.running && c.cancelConsume != nil {
		log.Println("Consumer: pausing consumption")
		c.cancelConsume()
		c.running = false
	}
}

// Resume снова подключает консьюмер к группе
func (c *Consumer) Resume(ctx context.Context) {
	c.consumeMutex.Lock()
	defer c.consumeMutex.Unlock()

	if !c.running {
		log.Println("Consumer: resuming consumption")
		consumeCtx, cancel := context.WithCancel(ctx)
		c.cancelConsume = cancel
		c.running = true

		go c.startConsumption(consumeCtx)
	}
}

func (c *Consumer) startConsumption(ctx context.Context) {
	handler := &consumerGroupHandler{
		messages: c.messages,
		closed:   c.closed,
	}

	retryDelay := time.Second * 5
	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer: context cancelled, stopping")
			return
		default:
			log.Println("Consumer: starting consumption cycle")
			err := c.group.Consume(ctx, []string{c.topic}, handler)
			if err != nil {
				log.Printf("Consume error: %v, retrying in %v", err, retryDelay)
				select {
				case <-ctx.Done():
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
}

// StartListening запускает асинхронное потребление сообщений
func (c *Consumer) StartListening(ctx context.Context) {
	c.consumeMutex.Lock()
	defer c.consumeMutex.Unlock()

	if !c.running {
		consumeCtx, cancel := context.WithCancel(ctx)
		c.cancelConsume = cancel
		c.running = true

		go c.startConsumption(consumeCtx)
	}
}

// Close останавливает потребитель и освобождает ресурсы
func (c *Consumer) Close() error {
	c.consumeMutex.Lock()
	defer c.consumeMutex.Unlock()

	if c.running && c.cancelConsume != nil {
		c.cancelConsume()
		c.running = false
	}
	close(c.closed)
	return c.group.Close()
}

// Messages возвращает канал для чтения сообщений
func (c *Consumer) Messages() <-chan consumerMessage {
	return c.messages
}

// consumerGroupHandler реализует интерфейс sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	messages chan<- consumerMessage
	closed   <-chan struct{}
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
			select {
			case h.messages <- consumerMessage{
				Value:   msg.Value,
				Session: sess,
				Message: msg,
			}:
				// Сообщение отправлено в канал, подтверждение будет после обработки
			case <-sess.Context().Done():
				return nil
			case <-h.closed:
				return nil
			}
		case <-sess.Context().Done():
			return nil
		case <-h.closed:
			return nil
		}
	}
}
