# Определяем операционную систему
ifeq ($(OS),Windows_NT)
    # Команда для Windows (PowerShell)
    CP = powershell -Command "if (-not (Test-Path .env)) { Copy-Item .env.example .env }"
else
    # Команда для Linux / macOS
    CP = cp -n .env.example .env
endif

.PHONY: env up down restart build

# 1. Автоматическое создание .env из шаблона и запуск проекта в фоне
up:
	@$(CP)
	docker compose --env-file .env -f deployments/docker-compose.yml up -d

# 2. Остановка контейнеров
down:
	docker compose -f deployments/docker-compose.yml down

# 3. Полная пересборка приложения (например, после изменения Go-кода или Dockerfile)
build:
	@$(CP)
	docker compose --env-file .env -f deployments/docker-compose.yml up --build -d

# 4. Перезапуск проекта
restart: down up
