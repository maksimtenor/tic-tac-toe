package main

import (
	"bohemas/internal/config"
	"bohemas/internal/database"
	"bohemas/internal/services"
	"bohemas/internal/transport/rest"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Configure()
	ctx := context.Background()
	redisDB := database.NewClient(cfg)

	if err := redisDB.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}
	log.Println("Redis connected successfully!")
	defer func(redis *redis.Client) {
		err := redis.Close()
		if err != nil {
			fmt.Printf("Redis connection close failed: %v", err)
		}
	}(redisDB)

	gameService := services.NewGameService(redisDB)
	mux := rest.NewRouter(gameService)

	fmt.Printf("Starting server on port %s\n", cfg.Port)
	fmt.Printf("Текст перед запуском... %s\n", cfg.Test)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
