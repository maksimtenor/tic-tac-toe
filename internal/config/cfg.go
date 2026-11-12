package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string `json:"port"`
	Test          string `json:"test"`
	RedisAddr     string `json:"redis_addr"`
	RedisUser     string `json:"redis_user"`
	RedisPassword string `json:"redis_password"`
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

	redisUser := os.Getenv("REDIS_USER")
	if redisUser == "" {
		redisUser = "default"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		redisPassword = ""
	}

	return &Config{
		Port:          port,
		Test:          os.Getenv("TEST"),
		RedisAddr:     redisHost + ":" + redisPort,
		RedisUser:     redisUser,
		RedisPassword: redisPassword,
	}
}
