package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string `json:"port"`
	Test string `json:"test"`
}

func Configure() *Config {
	// Пытаемся загрузить файл, но НЕ падаем, если его нет
	_ = godotenv.Load()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port: port,
		Test: os.Getenv("TEST"),
	}
}
