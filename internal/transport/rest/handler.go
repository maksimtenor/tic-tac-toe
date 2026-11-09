package rest

import (
	"bohemas/internal/app"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type SessionRequest struct {
	SessionID string `json:"sessionId"`
}

type MoveRequest struct {
	SessionID string `json:"sessionId"`
	Position  string `json:"position"`
}

type GameSession struct {
	Name     string
	UserSide string
	BotSide  string
	Builder  *app.BoardBuilder
	GameOver bool
	Mu       sync.Mutex
}

type StartRequest struct {
	Login string `json:"login"`
	Side  string `json:"side"`
}

var sessions = make(map[string]*GameSession)
var sessionsMu sync.Mutex

func newSessionID() (string, error) {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func Move(w http.ResponseWriter, r *http.Request) {
	randomKey := ""
	gameWinner := ""
	isDrawResponse := false
	var req MoveRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	sessionsMu.Lock()
	game, exists := sessions[req.SessionID]
	sessionsMu.Unlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	game.Mu.Lock()
	defer game.Mu.Unlock()

	builder := game.Builder
	userSide := game.UserSide
	botSide := game.BotSide

	if game.GameOver {
		http.Error(w, "Game already finished", http.StatusConflict)
		return
	}

	value, exists := builder.Board[req.Position]
	if !exists {
		http.Error(w, "Invalid cell", http.StatusBadRequest)
		return
	}
	if value != builder.Unused {
		http.Error(w, "Cell already occupied", http.StatusBadRequest)
		return
	}

	builder.Board[req.Position] = app.Player1 // выбор юзера
	chars := strings.Split(req.Position, "")
	rowNumber, _ := strconv.Atoi(chars[1])
	builder.BoardTemplate[chars[0]][rowNumber] = userSide

	// Группировка инфы по сторонам пользователей
	players := []string{
		userSide,
		botSide,
	}

	if win, winner := app.IsWinner(builder, players); win {
		gameWinner = winner
	}

	emptyCells := 0

	for _, value := range builder.Board {
		if value == builder.Unused {
			emptyCells++
		}
	}

	if emptyCells == 0 && gameWinner == "" {
		isDrawResponse = true
	}

	if gameWinner == "" && !isDrawResponse {
		randomKey = app.BotMove(builder, botSide, userSide)

		if randomKey != "" {
			builder.Board[randomKey] = app.Player2
		}

		// Формирование стороны и выбора BOT
		chars2 := strings.Split(randomKey, "")
		rowNumber2, _ := strconv.Atoi(chars2[1])
		builder.BoardTemplate[chars2[0]][rowNumber2] = botSide

		if win, winner := app.IsWinner(builder, players); win {
			gameWinner = winner
		}
	}
	emptyCells = 0
	for _, value := range builder.Board {
		if value == builder.Unused {
			emptyCells++
		}
	}
	if emptyCells == 0 && gameWinner == "" {
		isDrawResponse = true
	}

	if gameWinner != "" || isDrawResponse {
		game.GameOver = true
	}

	w.Header().Set("Content-Type", "application/json")
	responseData := map[string]interface{}{
		"botMove": randomKey,      // Переменная, куда записался ход бота (например, "b2")
		"winner":  gameWinner,     // Переменная со значением "X", "0" или ""
		"isDraw":  isDrawResponse, // Булево значение true/false
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

	sessionsMu.Lock()
	game, exists := sessions[req.SessionID]
	sessionsMu.Unlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	game.Mu.Lock()
	defer game.Mu.Unlock()

	game.Builder = app.NewBoardBuilder()
	game.GameOver = false

	w.WriteHeader(http.StatusOK)
}

func Start(w http.ResponseWriter, r *http.Request) {
	var req StartRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Login) == "" ||
		(req.Side != "X" && req.Side != "O") {
		http.Error(w, "Invalid player data", http.StatusBadRequest)
		return
	}

	botSide := "O"
	if req.Side == "O" {
		botSide = "X"
	}

	id, err := newSessionID()
	if err != nil {
		http.Error(w, "Session creation failed", http.StatusInternalServerError)
		return
	}

	game := &GameSession{
		Name:     strings.TrimSpace(req.Login),
		UserSide: req.Side,
		BotSide:  botSide,
		Builder:  app.NewBoardBuilder(),
	}

	sessionsMu.Lock()
	sessions[id] = game
	sessionsMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"sessionId": id,
		"userSide":  game.UserSide,
		"botSide":   game.BotSide,
	})
}
