package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string `json:"port"`
	Test      string `json:"test"`
	RedisAddr string `json:"redis_addr"`
}

func Configure() *Config {
	_ = godotenv.Load()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost" // Дефолт для локального запуска без докера
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	return &Config{
		Port:      port,
		Test:      os.Getenv("TEST"),
		RedisAddr: redisHost + ":" + redisPort,
	}
}
