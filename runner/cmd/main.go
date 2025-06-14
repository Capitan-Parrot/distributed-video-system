package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/config"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/database"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/kafka"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/runner"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/s3"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/services/detection"
)

func main() {
	log.Println("Main: init...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse cfg
	cfg, err := config.LoadConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация базы данных
	db, err := database.New(cfg.Postgres.DSN)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Init S3 client
	s3Client, err := s3.NewMinioClient(cfg.Minio.Endpoint, cfg.Minio.AccessKey, cfg.Minio.SecretKey)
	if err != nil {
		log.Fatalf("Failed to connect MinIO: %v", err)
	}

	// Start Kafka consumer for video processing
	consumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.ScenarioTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()
	go consumer.StartListening(ctx)

	// Start Kafka producer for heartbeats
	producer, err := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.HeartbeatTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	detectClient := detection.NewClient(cfg.Detection.Endpoint)

	r := runner.New(db, s3Client, detectClient, consumer, producer)
	go r.ListenAndRun(ctx)

	go r.ProcessStopEvent(ctx)

	// Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	cancel()
	log.Println("Main: Shutting down...")
}
