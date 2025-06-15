package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/models"
	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewProducer создаёт продюсер с настройками
func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *Producer) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka producer: %w", err)
	}
	return nil
}

// SendHeartbeat отправляет одно сообщение в Kafka
func (p *Producer) SendHeartbeat(msg models.Heartbeat) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	kafkaMsg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(msg.ScenarioID),
		Value: sarama.ByteEncoder(payload),
	}

	_, _, err = p.producer.SendMessage(kafkaMsg)
	if err != nil {
		return err
	}

	//log.Printf("Sent message to Kafka topic=%s partition=%d offset=%d", p.topic, partition, offset)
	return nil
}
