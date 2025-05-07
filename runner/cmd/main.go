package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/kafka"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/runner"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/s3"
	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/services/detection"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Init S3 client
	s3Client, err := s3.NewMinioClient("localhost:9000", "minio-access-key", "minio-secret-key")
	if err != nil {
		log.Fatalf("Ошибка подключения к MinIO: %v", err)
	}

	// Start Kafka consumer
	consumer, err := kafka.NewConsumer([]string{"localhost:9091"}, "video-runner-group", "video-scenarios")
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()
	go consumer.StartListening(ctx)

	detectClient := detection.NewClient("http://localhost:8000/predict")

	r := runner.New(s3Client, detectClient, consumer)
	go r.ListenAndRun(ctx)

	// Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("Завершение работы...")
	cancel() // Stop goroutines
}
