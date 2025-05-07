package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/api"
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/outbox"
	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/s3"
	"github.com/gorilla/mux"

	"github.com/Capitan-Parrot/distributed-video-system/orhestrator/internal/database"
)

func main() {
	// Инициализация базы данных
	dsn := "host=localhost port=5432 user=vidanalytics password=secret dbname=vidanalytics sslmode=disable"
	db, err := database.New(dsn)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Инициализация s3
	minioClient, err := s3.NewMinioClient("localhost:9000", "minio-access-key", "minio-secret-key")
	if err != nil {
		log.Fatalf("Не удалось подключиться к MinIO: %v", err)
	}

	// Горутина для обработки аутбокса
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go outbox.StartOutboxDispatcher(ctx, db, []string{"localhost:9091"}, "video-scenarios", 5*time.Second)

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
