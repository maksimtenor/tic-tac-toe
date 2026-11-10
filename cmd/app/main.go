package main

import (
	"bohemas/internal/config"
	"bohemas/internal/transport/rest"
	"fmt"
	"log"
	"net/http"
)

func main() {
	cfg := config.Configure()

	mux := rest.NewRouter()
	fmt.Printf("Starting server on port %s\n", cfg.Port)
	fmt.Printf("Текст перед запуском... %s\n", cfg.Test)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
