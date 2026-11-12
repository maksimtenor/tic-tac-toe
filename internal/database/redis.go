package database

import (
	"bohemas/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewClient(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Username: cfg.RedisUser,
		Password: cfg.RedisPassword,
		DB:       0,
	})
}
