package app

import (
	"bohemas/internal/app/helpers"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type MoveResult struct {
	BotMove string
	Winner  string
	IsDraw  bool
}

type ResetResult struct {
	Builder  *helpers.BoardBuilder
	GameOver bool
}

type StartResult struct {
	SessionId string
	Game      *GameSession
}

func PlayStart(login, side string) (StartResult, error) {
	if strings.TrimSpace(login) == "" ||
		(side != "X" && side != "O") {
		err := errors.New("invalid login")
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

	if win, winner := helpers.IsWinner(game.Builder, players); win {
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

		if win, winner := helpers.IsWinner(game.Builder, players); win {
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

func BotMove(builder *helpers.BoardBuilder, botSide, userSide string) string {
	// Проверяем, можно ли завершить линию за один ход
	findWinningMove := func(side string) string {
		for _, line := range helpers.Wins {
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
