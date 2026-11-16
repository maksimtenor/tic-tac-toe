package app

import (
	"testing"
)

func TestIsWinnerTable(t *testing.T) {
	tests := []struct {
		name       string
		board      map[string]map[int]string
		wantWin    bool
		wantWinner string
	}{
		{
			name: "Победа X по горизонтали",
			board: map[string]map[int]string{
				"a": {1: "X", 2: "X", 3: "X"},
				"b": {1: "_", 2: "_", 3: "_"},
				"c": {1: "_", 2: "_", 3: "_"},
			},
			wantWin:    true,
			wantWinner: "X",
		},
		{
			name: "Нет победителя",
			board: map[string]map[int]string{
				"a": {1: "X", 2: "O", 3: "X"},
				"b": {1: "_", 2: "_", 3: "O"},
				"c": {1: "O", 2: "X", 3: "_"},
			},
			wantWin:    false,
			wantWinner: "",
		},
		{
			name: "Победа O по диагонали",
			board: map[string]map[int]string{
				"a": {1: "O", 2: "X", 3: "O"},
				"b": {1: "_", 2: "O", 3: "X"},
				"c": {1: "X", 2: "X", 3: "O"},
			},
			wantWin:    true,
			wantWinner: "O",
		},
		{
			name: "Победа X по вертикали",
			board: map[string]map[int]string{
				"a": {1: "O", 2: "X", 3: "X"},
				"b": {1: "X", 2: "X", 3: "O"},
				"c": {1: "O", 2: "X", 3: "_"},
			},
			wantWin:    true,
			wantWinner: "X",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := NewBoardBuilder()
			board.BoardTemplate = tt.board

			gotWin, gotWinner := IsWinner(board, []string{"X", "O"})

			if gotWin != tt.wantWin {
				t.Errorf("got %v, want %v", gotWin, tt.wantWin)
			}

			if gotWinner != tt.wantWinner {
				t.Errorf("got %q, want %q", gotWinner, tt.wantWinner)
			}
		})
	}
}

func TestBotMoveTable(t *testing.T) {
	tests := []struct {
		name          string
		wantStep      string
		botSide       string
		userSide      string
		boardTemplate map[string]map[int]string
		board         map[string]int
	}{
		{
			name: "Бот может победить",
			boardTemplate: map[string]map[int]string{
				"a": {1: "O", 2: "O", 3: "_"},
				"b": {1: "_", 2: "X", 3: "X"},
				"c": {1: "_", 2: "_", 3: "_"},
			},
			botSide:  "O",
			userSide: "X",
			wantStep: "a3",
		},
		{
			name: "Бот должен защититься",
			boardTemplate: map[string]map[int]string{
				"a": {1: "X", 2: "X", 3: "_"},
				"b": {1: "O", 2: "_", 3: "_"},
				"c": {1: "O", 2: "X", 3: "_"},
			},
			botSide:  "O",
			userSide: "X",
			wantStep: "a3",
		},
		{
			name: "Бот может выиграть одним ходом",
			boardTemplate: map[string]map[int]string{
				"a": {1: "O", 2: "O", 3: "_"},
				"b": {1: "X", 2: "X", 3: "_"},
				"c": {1: "_", 2: "_", 3: "_"},
			},
			botSide:  "O",
			userSide: "X",
			wantStep: "a3",
		},
		{
			name: "Проверка случайного хода бота",
			boardTemplate: map[string]map[int]string{
				"a": {1: "X", 2: "O", 3: "X"},
				"b": {1: "O", 2: "X", 3: "O"},
				"c": {1: "_", 2: "X", 3: "O"},
			},
			board: map[string]int{
				"a1": 1, "a2": 2, "a3": 1,
				"b1": 2, "b2": 1, "b3": 2,
				"c1": 0, "c2": 1, "c3": 2,
			},
			botSide:  "O",
			userSide: "X",
			wantStep: "c1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := NewBoardBuilder()
			board.BoardTemplate = tt.boardTemplate
			if tt.board != nil {
				board.Board = tt.board
			}
			botMove := BotMove(board, tt.botSide, tt.userSide)
			if botMove != tt.wantStep {
				t.Errorf("got %v, want %v", botMove, tt.wantStep)
			}
		})
	}
}
