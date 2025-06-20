.PHONY: install run dev lint format docker-build docker-run

# Установка зависимостей
install:
	poetry install

# Запуск сервера
run:
	poetry run uvicorn gateway.main:app --host 0.0.0.0 --port 8080

# Запуск в dev-режиме с автоперезапуском
dev:
	poetry run uvicorn gateway.main:app --reload --host 0.0.0.0 --port 8080

# Проверка кода
lint:
	poetry run flake8 gateway

# Форматирование кода
format:
	poetry run black gateway

# Сборка Docker-образа
docker-build:
	docker build -t gateway .

# Запуск контейнера
docker-run:
	docker run -p 8000:8000 gateway -d
