package rest

import "net/http"

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", Home)
	mux.HandleFunc("GET /health", Health)
	mux.HandleFunc("POST /api/reset", Reset)
	mux.HandleFunc("POST /api/start", Start)
	mux.HandleFunc("POST /api/move", Move)

	return mux
}
