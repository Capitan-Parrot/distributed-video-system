package kafka

import (
	"encoding/json"
	"log"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/models"
	"github.com/IBM/sarama"
)

type Producer struct {
	Producer sarama.SyncProducer
	Topic    string
}

// NewKafkaProducer создаёт продюсер с настройками
func NewKafkaProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		Producer: producer,
		Topic:    topic,
	}, nil
}

// SendOutboxMessageToKafka отправляет одно сообщение в Kafka
func (kp *Producer) SendOutboxMessageToKafka(msg *models.OutboxMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	kafkaMsg := &sarama.ProducerMessage{
		Topic: kp.Topic,
		Key:   sarama.StringEncoder(msg.ScenarioID),
		Value: sarama.ByteEncoder(payload),
	}

	partition, offset, err := kp.Producer.SendMessage(kafkaMsg)
	if err != nil {
		return err
	}

	log.Printf("Sent message to Kafka topic=%s partition=%d offset=%d", kp.Topic, partition, offset)
	return nil
}
