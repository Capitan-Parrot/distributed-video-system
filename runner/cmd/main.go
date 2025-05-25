package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/config"
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

	// Init S3 client
	s3Client, err := s3.NewMinioClient(cfg.Minio.Endpoint, cfg.Minio.AccessKey, cfg.Minio.SecretKey)
	if err != nil {
		log.Fatalf("Failed to connect MinIO: %v", err)
	}

	// Start Kafka consumer
	consumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.Topic)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()
	go consumer.StartListening(ctx)

	detectClient := detection.NewClient(cfg.Detection.Endpoint)

	r := runner.New(s3Client, detectClient, consumer)
	go r.ListenAndRun(ctx)

	// Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	cancel()
	log.Println("Main: Shutting down...")
}
