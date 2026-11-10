ifeq ($(OS),Windows_NT)
    CP = powershell -Command "if (-not (Test-Path .env)) { Copy-Item .env.example .env }"
else
    CP = cp -n .env.example .env
endif

.PHONY: env up down restart build

# Запуск проекта
up:
	@$(CP)
	docker compose --env-file .env -f deployments/docker-compose.yml up -d

# !!! ДОБАВИЛИ ФЛАГ СЮДА !!!
down:
	docker compose --env-file .env -f deployments/docker-compose.yml down

# Сборка и запуск
build:
	@$(CP)
	docker compose --env-file .env -f deployments/docker-compose.yml up --build -d

# Перезапуск
restart: down up
