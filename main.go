package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type BoardBuilder struct {
	unused        int
	rows          []string
	cols          []int
	sides         []string
	board         map[string]int
	boardTemplate map[string]map[int]string
}

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
	Builder  *BoardBuilder
	GameOver bool
	Mu       sync.Mutex
}

type StartRequest struct {
	Login string `json:"login"`
	Side  string `json:"side"`
}

var sessions = make(map[string]*GameSession)
var sessionsMu sync.Mutex
var player1, player2, unused = 1, 2, 0

func newBoardBuilder() *BoardBuilder {
	return &BoardBuilder{
		unused,
		[]string{"a", "b", "c"},
		[]int{1, 2, 3},
		[]string{"X", "O"},
		map[string]int{
			"a1": unused, "a2": unused, "a3": unused,
			"b1": unused, "b2": unused, "b3": unused,
			"c1": unused, "c2": unused, "c3": unused,
		},
		map[string]map[int]string{
			"a": {1: "_", 2: "_", 3: "_"},
			"b": {1: "_", 2: "_", 3: "_"},
			"c": {1: "_", 2: "_", 3: "_"},
		},
	}
}

func newSessionID() (string, error) {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// Запуск скрипта
func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("POST /api/reset", func(w http.ResponseWriter, r *http.Request) {
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

		game.Builder = newBoardBuilder()
		game.GameOver = false

		w.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("POST /api/start", func(w http.ResponseWriter, r *http.Request) {
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
			Builder:  newBoardBuilder(),
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
	})
	http.HandleFunc("POST /api/move", func(w http.ResponseWriter, r *http.Request) {
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

		value, exists := builder.board[req.Position]
		if !exists {
			http.Error(w, "Invalid cell", http.StatusBadRequest)
			return
		}
		if value != builder.unused {
			http.Error(w, "Cell already occupied", http.StatusBadRequest)
			return
		}

		builder.board[req.Position] = player1 // выбор юзера
		chars := strings.Split(req.Position, "")
		rowNumber, _ := strconv.Atoi(chars[1])
		builder.boardTemplate[chars[0]][rowNumber] = userSide

		// Группировка инфы по сторонам пользователей
		players := []string{
			userSide,
			botSide,
		}

		if win, winner := isWinner(builder, players); win {
			gameWinner = winner
		}

		emptyCells := 0

		for _, value := range builder.board {
			if value == builder.unused {
				emptyCells++
			}
		}

		if emptyCells == 0 && gameWinner == "" {
			isDrawResponse = true
		}

		if gameWinner == "" && !isDrawResponse {
			randomKey = botMove(builder, botSide, userSide)

			if randomKey != "" {
				builder.board[randomKey] = player2
			}

			// Формирование стороны и выбора BOT
			chars2 := strings.Split(randomKey, "")
			rowNumber2, _ := strconv.Atoi(chars2[1])
			builder.boardTemplate[chars2[0]][rowNumber2] = botSide

			if win, winner := isWinner(builder, players); win {
				gameWinner = winner
			}
		}
		emptyCells = 0
		for _, value := range builder.board {
			if value == builder.unused {
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
	})

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	http.ListenAndServe(":"+port, nil)
}

// isWinner Проверка победителя
func isWinner(builder *BoardBuilder, players []string) (bool, string) {
	wg := sync.WaitGroup{}
	ch := make(chan string, len(players))

	wg.Add(len(players))
	for _, player := range players {
		go check(player, builder, ch, &wg)
	}
	wg.Wait()
	close(ch)

	playerWon, ok := <-ch
	if ok {
		return true, playerWon
	}

	return false, ""
}

func check(player string, builder *BoardBuilder, ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	board := builder.boardTemplate

	// Проверка 3 горизонталей
	if board["a"][1] == player && board["a"][2] == player && board["a"][3] == player {
		ch <- player
	} else if board["b"][1] == player && board["b"][2] == player && board["b"][3] == player {
		ch <- player
	} else if board["c"][1] == player && board["c"][2] == player && board["c"][3] == player {
		ch <- player
	} else if board["a"][1] == player && board["b"][1] == player && board["c"][1] == player {
		ch <- player
	} else if board["a"][2] == player && board["b"][2] == player && board["c"][2] == player {
		ch <- player
	} else if board["a"][3] == player && board["b"][3] == player && board["c"][3] == player {
		ch <- player
	} else if board["a"][1] == player && board["b"][2] == player && board["c"][3] == player {
		ch <- player
	} else if board["a"][3] == player && board["b"][2] == player && board["c"][1] == player {
		ch <- player
	}
}

func botMove(builder *BoardBuilder, botSide, userSide string) string {
	// Все возможные выигрышные комбинации
	wins := [][]string{
		{"a1", "a2", "a3"},
		{"b1", "b2", "b3"},
		{"c1", "c2", "c3"},
		{"a1", "b1", "c1"},
		{"a2", "b2", "c2"},
		{"a3", "b3", "c3"},
		{"a1", "b2", "c3"},
		{"a3", "b2", "c1"},
	}

	// Проверяем, можно ли завершить линию за один ход
	findWinningMove := func(side string) string {
		for _, line := range wins {
			count := 0
			empty := ""

			for _, cell := range line {
				row := string(cell[0])
				col := int(cell[1] - '0')

				value := builder.boardTemplate[row][col]

				if value == side {
					count++
				} else if value == "_" {
					empty = cell
				}
			}

			if count == 2 && empty != "" {
				return empty
			}
		}

		return ""
	}

	// 1. Пытаемся победить
	if move := findWinningMove(botSide); move != "" {
		return move
	}

	// 2. Блокируем победу игрока
	if move := findWinningMove(userSide); move != "" {
		return move
	}

	// 3. Собираем свободные клетки
	freeCells := []string{}

	for cell, value := range builder.board {
		if value == builder.unused {
			freeCells = append(freeCells, cell)
		}
	}

	if len(freeCells) == 0 {
		return ""
	}

	// 4. Выбираем случайную свободную клетку
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(freeCells))))
	if err != nil {
		return freeCells[0]
	}

	return freeCells[n.Int64()]
}
