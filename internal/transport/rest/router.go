package rest

import (
	"bohemas/internal/services"
	"net/http"
)

func NewRouter(gameService *services.GameService) *http.ServeMux {
	mux := http.NewServeMux()

	handler := NewHandler(gameService)

	mux.HandleFunc("GET /", Home)
	mux.HandleFunc("GET /health", Health)
	mux.HandleFunc("POST /api/reset", handler.Reset)
	mux.HandleFunc("POST /api/start", handler.Start)
	mux.HandleFunc("POST /api/move", handler.Move)
	mux.HandleFunc("GET /api/session", handler.Session)

	return mux
}
