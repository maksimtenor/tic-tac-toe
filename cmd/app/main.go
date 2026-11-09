package main

import (
	"bohemas/internal/transport/rest"
	"log"
	"net/http"
)

func main() {
	mux := rest.NewRouter()

	log.Fatal(http.ListenAndServe(":8080", mux))
}
