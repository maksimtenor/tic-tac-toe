package app

import (
	"crypto/rand"
	"math/big"
	"sync"
)

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

// Все возможные выигрышные комбинации
var Wins = [][]string{
	{"a1", "a2", "a3"},
	{"b1", "b2", "b3"},
	{"c1", "c2", "c3"},
	{"a1", "b1", "c1"},
	{"a2", "b2", "c2"},
	{"a3", "b3", "c3"},
	{"a1", "b2", "c3"},
	{"a3", "b2", "c1"},
}

func BotMove(builder *BoardBuilder, botSide, userSide string) string {
	// Проверяем, можно ли завершить линию за один ход
	findWinningMove := func(side string) string {
		for _, line := range Wins {
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

func IsWinner(builder *BoardBuilder, players []string) (bool, string) {
	wg := sync.WaitGroup{}
	ch := make(chan string, len(players))

	wg.Add(len(players))
	for _, player := range players {
		go CheckWinner(player, builder, ch, &wg)
	}
	wg.Wait()
	close(ch)

	playerWon, ok := <-ch
	if ok {
		return true, playerWon
	}

	return false, ""
}

func CheckWinner(player string, builder *BoardBuilder, ch chan string, wg *sync.WaitGroup) {
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
