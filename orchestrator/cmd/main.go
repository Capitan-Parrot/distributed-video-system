package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	watchdog "github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal"
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/api"
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/config"
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/kafka"
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/outbox"
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/s3"
	"github.com/gorilla/mux"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/database"
)

func main() {
	log.Println("Main: init...")

	// Чтение конфига
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

	// Инициализация s3
	minioClient, err := s3.NewMinioClient(cfg.Minio.Endpoint, cfg.Minio.AccessKey, cfg.Minio.SecretKey)
	if err != nil {
		log.Fatalf("Failed connect to MinIO: %v", err)
	}

	// Горутина для обработки аутбокса
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go outbox.StartOutboxDispatcher(ctx, db, cfg.Kafka.Brokers, cfg.Kafka.ScenarioTopic, 5*time.Second)

	// Горутина для обработки heartbeats раннера
	consumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.HeartbeatTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()
	go consumer.StartListening(ctx, db)

	// Горутина для перезапуска упавших раннеров
	watchDog := watchdog.New(db)
	go watchDog.Start(ctx)

	// Настройка роутера
	r := mux.NewRouter()
	handlers := api.NewHandlers(db, minioClient)

	// Регистрация обработчиков
	r.HandleFunc("/scenario", handlers.CreateScenarioHandler).Methods("POST")
	r.HandleFunc("/scenario/{scenario_id}", handlers.GetScenarioStatusHandler).Methods("GET")
	r.HandleFunc("/scenario/{scenario_id}", handlers.UpdateScenarioStatusHandler).Methods("POST")
	r.HandleFunc("/prediction/{scenario_id}", handlers.GetPredictionsHandler).Methods("GET")

	// Запуск сервера
	log.Println("Starting orchestrator API server on :8002")
	log.Fatal(http.ListenAndServe(":8002", r))
}
