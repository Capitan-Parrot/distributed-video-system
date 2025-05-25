# Стадия сборки
FROM golang:1.24-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальные файлы проекта
COPY . .

# Сборка приложения
RUN go build -o orchestrator ./cmd

# Финальный образ
FROM alpine:latest


# Устанавливаем сертификаты (для HTTPS запросов и т.д.)
RUN apk --no-cache add ca-certificates

# Устанавливаем рабочую директорию
WORKDIR /root/

# Копируем собранный бинарник из стадии сборки
COPY --from=builder /app/orchestrator .

# Копируем конфигурацию
COPY --from=builder /app/internal/config ./internal/config

# Указываем порт, который слушает приложение
EXPOSE 8002

# Запуск приложения
CMD ["./orchestrator"]
