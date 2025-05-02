.PHONY: install run dev lint format docker-build docker-run

# Установка зависимостей
install:
	poetry install

# Запуск сервера
run:
	poetry run uvicorn detection.main:app --host 0.0.0.0 --port 8080

# Запуск в dev-режиме с автоперезапуском
dev:
	poetry run uvicorn detection.main:app --reload --host 0.0.0.0 --port 8080

# Проверка кода
lint:
	poetry run flake8 app

# Форматирование кода
format:
	poetry run black app

# Сборка Docker-образа
docker-build:
	docker build -t detection .

# Запуск контейнера
docker-run:
	docker run -d -p 8004:8004 detection
