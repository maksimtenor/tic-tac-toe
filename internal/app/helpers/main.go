package helpers

import "sync"

type BoardBuilder struct {
	Unused        int
	Rows          []string
	Cols          []int
	Sides         []string
	Board         map[string]int
	BoardTemplate map[string]map[int]string
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
