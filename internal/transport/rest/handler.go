package rest

import (
	"bohemas/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
)

type SessionRequest struct {
	SessionID string `json:"sessionId"`
}

type MoveRequest struct {
	Position string `json:"position"`
}

type StartRequest struct {
	Login string `json:"login"`
	Side  string `json:"side"`
}

type Handler struct {
	gameService *services.GameService
}

func NewHandler(gameService *services.GameService) *Handler {
	return &Handler{
		gameService: gameService,
	}
}

func (h *Handler) Move(w http.ResponseWriter, r *http.Request) {
	var req MoveRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("ttt_session")
	if err != nil {
		http.Error(w, "Session cookie not found", http.StatusUnauthorized)
		return
	}

	sessionID := cookie.Value

	game, err := h.gameService.PlayMove(r.Context(), sessionID, req.Position)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responseData := map[string]interface{}{
		"botMove": game.BotMove, // Переменная, куда записался ход бота (например, "b2")
		"winner":  game.Winner,  // Переменная со значением "X", "0" или ""
		"isDraw":  game.IsDraw,  // Булево значение true/false
	}
	json.NewEncoder(w).Encode(responseData)

	return
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func Home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/index.html")
}

func (h *Handler) Reset(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("ttt_session")
	if err != nil {
		http.Error(w, "Session cookie not found", http.StatusUnauthorized)
		return
	}

	sessionID := cookie.Value

	err = h.gameService.PlayReset(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Ошибка при сбросе игрового процесса", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	var req StartRequest
	var login, side string

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	login = req.Login
	side = req.Side

	start, err := h.gameService.PlayStart(r.Context(), login, side)
	if err != nil {
		http.Error(w, "Ошибка при запуске игры", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "ttt_session",
		Value:    start.SessionId,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3000,
		Secure:   false,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"sessionId": start.SessionId,
		"userSide":  start.Game.UserSide,
		"botSide":   start.Game.BotSide,
	})
}

func (h *Handler) Session(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("ttt_session")
	if err != nil {
		http.Error(w, "Session cookie not found", http.StatusUnauthorized)
		return
	}

	sessionID := cookie.Value
	gameSession, err := h.gameService.GetSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"userSide":  gameSession.UserSide,
		"botSide":   gameSession.BotSide,
		"sessionId": sessionID,
		"name":      gameSession.Name,
		"board":     gameSession.Builder.BoardTemplate,
		"gameOver":  gameSession.GameOver,
	})
}
