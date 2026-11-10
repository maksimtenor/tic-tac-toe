package rest

import (
	"bohemas/internal/app"
	"encoding/json"
	"fmt"
	"net/http"
)

type SessionRequest struct {
	SessionID string `json:"sessionId"`
}

type MoveRequest struct {
	SessionID string `json:"sessionId"`
	Position  string `json:"position"`
}

type StartRequest struct {
	Login string `json:"login"`
	Side  string `json:"side"`
}

func Move(w http.ResponseWriter, r *http.Request) {
	var req MoveRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	game, err := app.PlayMove(req.SessionID, req.Position)
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

func Reset(w http.ResponseWriter, r *http.Request) {
	var req SessionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := app.PlayReset(req.SessionID)
	if err != nil {
		http.Error(w, "Ошибка при сбросе игрового процесса", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func Start(w http.ResponseWriter, r *http.Request) {
	var req StartRequest
	var login, side string

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	login = req.Login
	side = req.Side

	start, err := app.PlayStart(login, side)
	if err != nil {
		http.Error(w, "Ошибка при запуске игры", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"sessionId": start.SessionId,
		"userSide":  start.Game.UserSide,
		"botSide":   start.Game.BotSide,
	})
}
