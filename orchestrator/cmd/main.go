package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/api"
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/config"
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
	go outbox.StartOutboxDispatcher(ctx, db, cfg.Kafka.Brokers, cfg.Kafka.Topic, 5*time.Second)

	// Настройка роутера
	r := mux.NewRouter()
	handlers := api.NewHandlers(db, minioClient)

	// Регистрация обработчиков
	r.HandleFunc("/scenario", handlers.CreateScenarioHandler).Methods("POST")
	r.HandleFunc("/scenario/{scenario_id}", handlers.UpdateScenarioStatusHandler).Methods("POST")
	r.HandleFunc("/scenario/{scenario_id}", handlers.GetScenarioStatusHandler).Methods("GET")
	r.HandleFunc("/prediction/{scenario_id}", handlers.GetPredictionsHandler).Methods("GET")

	// Запуск сервера
	log.Println("Starting orchestrator API server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
