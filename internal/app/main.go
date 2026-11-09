package app

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
)

type BoardBuilder struct {
	Unused        int
	Rows          []string
	Cols          []int
	Sides         []string
	Board         map[string]int
	BoardTemplate map[string]map[int]string
}

var Player1, Player2, unused = 1, 2, 0

func NewBoardBuilder() *BoardBuilder {
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

type GameSession struct {
	Name     string
	UserSide string
	BotSide  string
	Builder  *BoardBuilder
	GameOver bool
	Mu       sync.Mutex
}

var sessions = make(map[string]*GameSession)
var sessionsMu sync.Mutex

type MoveResult struct {
	BotMove string
	Winner  string
	IsDraw  bool
}

type ResetResult struct {
	Builder  *BoardBuilder
	GameOver bool
}

type StartResult struct {
	SessionId string
	Game      *GameSession
}

func newSessionID() (string, error) {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func PlayStart(login, side string) (StartResult, error) {
	if strings.TrimSpace(login) == "" ||
		(side != "X" && side != "O") {
		err := errors.New("Invalid login")
		return StartResult{}, err
	}

	botSide := "O"
	if side == "O" {
		botSide = "X"
	}

	id, err := newSessionID()
	if err != nil {
		err = errors.New(fmt.Sprint("Error creating new session ID:", err))
		return StartResult{}, err
	}

	game := &GameSession{
		Name:     strings.TrimSpace(login),
		UserSide: side,
		BotSide:  botSide,
		Builder:  NewBoardBuilder(),
	}

	start := StartResult{
		SessionId: id,
		Game:      game,
	}

	sessionsMu.Lock()
	sessions[id] = game
	sessionsMu.Unlock()

	return start, nil
}

func PlayReset(sessionId string) error {
	sessionsMu.Lock()
	game, exists := sessions[sessionId]
	sessionsMu.Unlock()

	if !exists {
		err := errors.New("session not found")
		return err
	}

	game.Mu.Lock()
	defer game.Mu.Unlock()

	game.Builder = NewBoardBuilder()
	game.GameOver = false

	return nil
}

func PlayMove(sessionId string, position string) (MoveResult, error) {
	move := MoveResult{
		BotMove: "",
		Winner:  "",
		IsDraw:  false,
	}

	sessionsMu.Lock()
	game, exists := sessions[sessionId]
	sessionsMu.Unlock()

	if !exists {
		err := errors.New("session not found")
		return MoveResult{}, err
	}

	game.Mu.Lock()
	defer game.Mu.Unlock()

	if game.GameOver {
		err := errors.New("game already finished")
		return MoveResult{}, err
	}

	value, exists := game.Builder.Board[position]
	if !exists {
		err := errors.New("Invalid cell")
		return MoveResult{}, err
	}
	if value != game.Builder.Unused {
		err := errors.New("Cell already occupied")
		return MoveResult{}, err
	}

	game.Builder.Board[position] = Player1 // выбор юзера
	chars := strings.Split(position, "")
	rowNumber, _ := strconv.Atoi(chars[1])
	game.Builder.BoardTemplate[chars[0]][rowNumber] = game.UserSide

	// Группировка инфы по сторонам пользователей
	players := []string{
		game.UserSide,
		game.BotSide,
	}

	if win, winner := IsWinner(game.Builder, players); win {
		move.Winner = winner
	}

	emptyCells := 0

	for _, value := range game.Builder.Board {
		if value == game.Builder.Unused {
			emptyCells++
		}
	}

	if emptyCells == 0 && move.Winner == "" {
		move.IsDraw = true
	}

	if move.Winner == "" && !move.IsDraw {
		move.BotMove = BotMove(game.Builder, game.BotSide, game.UserSide)

		if move.BotMove != "" {
			game.Builder.Board[move.BotMove] = Player2
		}

		// Формирование стороны и выбора BOT
		chars2 := strings.Split(move.BotMove, "")
		rowNumber2, _ := strconv.Atoi(chars2[1])
		game.Builder.BoardTemplate[chars2[0]][rowNumber2] = game.BotSide

		if win, winner := IsWinner(game.Builder, players); win {
			move.Winner = winner
		}
	}
	emptyCells = 0
	for _, value := range game.Builder.Board {
		if value == game.Builder.Unused {
			emptyCells++
		}
	}
	if emptyCells == 0 && move.Winner == "" {
		move.IsDraw = true
	}

	if move.Winner != "" || move.IsDraw {
		game.GameOver = true
	}

	return move, nil
}

// isWinner Проверка победителя
func IsWinner(builder *BoardBuilder, players []string) (bool, string) {
	wg := sync.WaitGroup{}
	ch := make(chan string, len(players))

	wg.Add(len(players))
	for _, player := range players {
		go Check(player, builder, ch, &wg)
	}
	wg.Wait()
	close(ch)

	playerWon, ok := <-ch
	if ok {
		return true, playerWon
	}

	return false, ""
}

func Check(player string, builder *BoardBuilder, ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	board := builder.BoardTemplate

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

func BotMove(builder *BoardBuilder, botSide, userSide string) string {
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

				value := builder.BoardTemplate[row][col]

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

	for cell, value := range builder.Board {
		if value == builder.Unused {
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
