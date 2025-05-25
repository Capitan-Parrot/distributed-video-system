COMPOSE_FILE=docker-compose.yaml
COMPOSE_BAKE=true

.PHONY: up down restart logs ps build

up:
	docker-compose -f $(COMPOSE_FILE) up -d

down:
	docker-compose -f $(COMPOSE_FILE) down

restart:
	docker-compose -f $(COMPOSE_FILE) down
	docker-compose -f $(COMPOSE_FILE) up -d --build

build:
	docker-compose -f $(COMPOSE_FILE) build

dependencies:
	docker-compose up -d postgres kafka zookeeper kafka-ui minio

