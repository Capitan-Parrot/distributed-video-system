package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"gopkg.in/yaml.v3"
)

// Config структура конфига
type Config struct {
	Postgres struct {
		DSN string `yaml:"dsn" env:"DATABASE_DSN"`
	} `yaml:"postgres"`

	Minio struct {
		Endpoint  string `yaml:"endpoint" env:"MINIO_ENDPOINT"`
		AccessKey string `yaml:"access_key" env:"MINIO_ACCESS_KEY"`
		SecretKey string `yaml:"secret_key" env:"MINIO_SECRET_KEY"`
	} `yaml:"minio"`

	Kafka struct {
		Brokers        []string `yaml:"brokers" env:"KAFKA_BROKERS" envSeparator:","`
		GroupID        string   `yaml:"group_id" env:"KAFKA_GROUP_ID"`
		ScenarioTopic  string   `yaml:"scenario_topic" env:"SCENARIO_TOPIC"`
		HeartbeatTopic string   `yaml:"heartbeat_topic" env:"HEARTBEAT_TOPIC"`
	} `yaml:"kafka"`

	Detection struct {
		Endpoint string `yaml:"endpoint" env:"DETECTION_ENDPOINT"`
	} `yaml:"detection"`
}

func LoadConfig(filename string) (*Config, error) {
	cfg := &Config{}

	if filename == "" {
		filename = "local.yaml"
	}
	path := "internal/config/" + filename

	// Читаем YAML
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Парсим YAML в структуру
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Парсим переменные окружения с приоритетом
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	fmt.Println(cfg)
	return cfg, nil
}
