# === ЭТАП СБОРКИ (BUILD STAGE) ===
FROM golang:1.26.0-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости (они закэшируются Docker'ом)
RUN go mod download

# Копируем весь проект (cmd, internal, файлы конфигурации)
COPY . .

# Собираем проект из вашей директории cmd/app/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp ./cmd/app/main.go


# === ЭТАП ЗАПУСКА (RUN STAGE) ===
FROM alpine:latest

WORKDIR /app

# Устанавливаем таймзоны и сертификаты безопасности (полезно для Go)
RUN apk --no-cache add ca-certificates tzdata

# Копируем скомпилированный бинарник из контейнера-сборщика
COPY --from=builder /app/myapp .

# !!! КРИТИЧЕСКИЙ ШАГ: Копируем папку с фронтендом в финальный образ !!!
COPY --from=builder /app/web ./web

# Копируем файл окружения в корень приложения
COPY .env .

# Запуск приложения
CMD ["./myapp"]
