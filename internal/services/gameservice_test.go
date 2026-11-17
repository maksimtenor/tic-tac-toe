package services

import (
	"bohemas/internal/app"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

type FakeStore struct {
	data map[string]string
}

func (f *FakeStore) Get(ctx context.Context, key string) *redis.StringCmd {
	val, ok := f.data[key]
	if !ok {
		return redis.NewStringResult("", redis.Nil)
	}
	return redis.NewStringResult(val, nil)
}

func (f *FakeStore) Set(
	ctx context.Context,
	key string,
	value interface{},
	expiration time.Duration,
) *redis.StatusCmd {
	switch v := value.(type) {
	case string:
		f.data[key] = v
	case []byte:
		f.data[key] = string(v)
	default:
		return redis.NewStatusResult("", fmt.Errorf("unsupported value type: %T", value))
	}
	return redis.NewStatusResult("OK", nil)
}

func TestPlayMove(t *testing.T) {
	tests := []struct {
		name          string
		ctx           context.Context
		Name          string
		UserSide      string
		BotSide       string
		sessionID     string
		fakeData      map[string]string
		position      string
		wantTextError string
		wantErr       bool
	}{
		{
			name:      "Ход в свободную клетку",
			ctx:       context.Background(),
			Name:      "Test",
			UserSide:  "X",
			BotSide:   "O",
			sessionID: "12345",
			position:  "a3",
			fakeData: map[string]string{
				"game-session:12345": "{\"Name\":\"Test\",\"UserSide\":\"O\",\"BotSide\":\"X\",\"Builder\":{\"Unused\":0,\"Rows\":[\"a\",\"b\",\"c\"],\"Cols\":[1,2,3],\"Sides\":[\"X\",\"O\"],\"Board\":{\"a1\":2,\"a2\":0,\"a3\":0,\"b1\":0,\"b2\":1,\"b3\":0,\"c1\":0,\"c2\":0,\"c3\":0},\"BoardTemplate\":{\"a\":{\"1\":\"X\",\"2\":\"_\",\"3\":\"_\"},\"b\":{\"1\":\"_\",\"2\":\"O\",\"3\":\"_\"},\"c\":{\"1\":\"_\",\"2\":\"_\",\"3\":\"_\"}}},\"GameOver\":false}",
			},
			wantErr: false,
		},
		{
			name:      "Ход в занятую клетку",
			ctx:       context.Background(),
			Name:      "Test",
			UserSide:  "X",
			BotSide:   "O",
			sessionID: "12345",
			position:  "b2",
			fakeData: map[string]string{
				"game-session:12345": "{\"Name\":\"Test\",\"UserSide\":\"O\",\"BotSide\":\"X\",\"Builder\":{\"Unused\":0,\"Rows\":[\"a\",\"b\",\"c\"],\"Cols\":[1,2,3],\"Sides\":[\"X\",\"O\"],\"Board\":{\"a1\":2,\"a2\":0,\"a3\":0,\"b1\":0,\"b2\":1,\"b3\":0,\"c1\":0,\"c2\":0,\"c3\":0},\"BoardTemplate\":{\"a\":{\"1\":\"X\",\"2\":\"_\",\"3\":\"_\"},\"b\":{\"1\":\"_\",\"2\":\"O\",\"3\":\"_\"},\"c\":{\"1\":\"_\",\"2\":\"_\",\"3\":\"_\"}}},\"GameOver\":false}",
			},
			wantErr:       true,
			wantTextError: "cell already occupied",
		},
		{
			name:      "Несуществующая клетка z9",
			ctx:       context.Background(),
			Name:      "Test",
			UserSide:  "X",
			BotSide:   "O",
			sessionID: "12345",
			position:  "z9",
			fakeData: map[string]string{
				"game-session:12345": "{\"Name\":\"Test\",\"UserSide\":\"O\",\"BotSide\":\"X\",\"Builder\":{\"Unused\":0,\"Rows\":[\"a\",\"b\",\"c\"],\"Cols\":[1,2,3],\"Sides\":[\"X\",\"O\"],\"Board\":{\"a1\":2,\"a2\":0,\"a3\":0,\"b1\":0,\"b2\":1,\"b3\":0,\"c1\":0,\"c2\":0,\"c3\":0},\"BoardTemplate\":{\"a\":{\"1\":\"X\",\"2\":\"_\",\"3\":\"_\"},\"b\":{\"1\":\"_\",\"2\":\"O\",\"3\":\"_\"},\"c\":{\"1\":\"_\",\"2\":\"_\",\"3\":\"_\"}}},\"GameOver\":false}",
			},
			wantErr:       true,
			wantTextError: "invalid cell",
		},
		{
			name:      "Игра уже завершена",
			ctx:       context.Background(),
			Name:      "Test",
			UserSide:  "X",
			BotSide:   "O",
			sessionID: "12345",
			position:  "b1",
			fakeData: map[string]string{
				"game-session:12345": "{\"Name\":\"Test\",\"UserSide\":\"O\",\"BotSide\":\"X\",\"Builder\":{\"Unused\":0,\"Rows\":[\"a\",\"b\",\"c\"],\"Cols\":[1,2,3],\"Sides\":[\"X\",\"O\"],\"Board\":{\"a1\":2,\"a2\":1,\"a3\":1,\"b1\":0,\"b2\":1,\"b3\":1,\"c1\":2,\"c2\":2,\"c3\":2},\"BoardTemplate\":{\"a\":{\"1\":\"X\",\"2\":\"O\",\"3\":\"O\"},\"b\":{\"1\":\"_\",\"2\":\"O\",\"3\":\"O\"},\"c\":{\"1\":\"X\",\"2\":\"X\",\"3\":\"X\"}}},\"GameOver\":true}",
			},
			wantErr:       true,
			wantTextError: "game already finished",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.fakeData
			service := NewGameService(&FakeStore{
				data: data,
			})

			_, err := service.PlayMove(tt.ctx, tt.sessionID, tt.position)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if err != nil {
				if tt.wantErr && tt.wantTextError != err.Error() {
					t.Errorf("got = %v, want %v", err.Error(), tt.wantTextError)
				}
				return
			}
			session, err := service.GetSession(tt.ctx, tt.sessionID)
			if err != nil {
				t.Fatal(err)
			}

			if session.Builder.Board["a3"] != app.Player1 {
				t.Errorf("got = %v, want %v", session.Builder.Board["a3"], app.Player1)
			}
			if session.Builder.BoardTemplate["a"][3] != "O" {
				t.Errorf("got = %v, want %v", session.Builder.BoardTemplate["a"][3], "O")
			}
		})
	}
}

func TestPlayReset(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		Name     string
		UserSide string
		BotSide  string
	}{
		{
			name:     "Сброс завершённой игры",
			ctx:      context.Background(),
			Name:     "Test",
			UserSide: "X",
			BotSide:  "O",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeStore := &FakeStore{
				make(map[string]string),
			}
			service := NewGameService(fakeStore)

			playStart, err := service.PlayStart(tt.ctx, tt.Name, tt.UserSide)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			session, err := service.GetSession(tt.ctx, playStart.SessionId)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			session.GameOver = true
			session.Builder.Board["a1"] = app.Player1
			session.Builder.BoardTemplate["a"][1] = tt.UserSide

			jsonData, err := json.Marshal(session)
			if err != nil {
				t.Fatal(err)
			}

			err = fakeStore.Set(
				tt.ctx,
				"game-session:"+playStart.SessionId,
				jsonData,
				3000*time.Second,
			).Err()

			if err != nil {
				t.Fatal(err)
			}

			err = service.PlayReset(tt.ctx, playStart.SessionId)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			session, err = service.GetSession(tt.ctx, playStart.SessionId)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if session.GameOver {
				t.Errorf("Game over")
			}
			if session.Builder.Board["a1"] != 0 {
				t.Errorf("A1 not reset")
			}
			if session.Builder.BoardTemplate["a"][1] != "_" {
				t.Errorf("expected empty a1, got %q", session.Builder.BoardTemplate["a"][1])
			}
			if session.Name != tt.Name {
				t.Errorf("Name not reset")
			}
			if session.UserSide != tt.UserSide {
				t.Errorf("UserSide not reset")
			}
			if session.BotSide != tt.BotSide {
				t.Errorf("BotSide not reset")
			}
		})
	}
}

func TestPlayStart(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		Name     string
		UserSide string
		BotSide  string
		wantErr  bool
	}{
		{
			name:     "Тестируем создание сессии",
			ctx:      context.Background(),
			wantErr:  false,
			Name:     "Test",
			UserSide: "X",
			BotSide:  "O",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeStore := &FakeStore{
				make(map[string]string),
			}
			service := NewGameService(fakeStore)
			play, err := service.PlayStart(tt.ctx, tt.Name, tt.UserSide)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if play.SessionId == "" {
				t.Fatal("expected non-empty sessionID")
			}

			session, err := service.GetSession(tt.ctx, play.SessionId)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if session == nil {
				t.Fatal("expected session, got nil")
			}

			if session.Name != tt.Name {
				t.Errorf("got name %q, want %q", session.Name, tt.Name)
			}
			if session.UserSide != tt.UserSide {
				t.Errorf("got side %q, want %q", session.UserSide, tt.UserSide)
			}
			if session.BotSide != tt.BotSide {
				t.Errorf("got side %q, want %q", session.BotSide, tt.BotSide)
			}
		})
	}
}

func TestGetSession(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		fakeData    map[string]string
		sessionID   string
		wantErr     bool
		wantErrText string
		Name        string
		UserSide    string
	}{
		{
			name:      "Сессия существует → возвращается корректная GameSession",
			ctx:       context.Background(),
			sessionID: "12345",
			fakeData: map[string]string{
				"game-session:12345": "{\"Name\":\"Test\",\"UserSide\":\"X\",\"BotSide\":\"O\",\"Builder\":{\"Unused\":0,\"Rows\":[\"a\",\"b\",\"c\"],\"Cols\":[1,2,3],\"Sides\":[\"X\",\"O\"],\"Board\":{\"a1\":1,\"a2\":2,\"a3\":1,\"b1\":2,\"b2\":1,\"b3\":1,\"c1\":2,\"c2\":1,\"c3\":2},\"BoardTemplate\":{\"a\":{\"1\":\"X\",\"2\":\"O\",\"3\":\"X\"},\"b\":{\"1\":\"O\",\"2\":\"X\",\"3\":\"X\"},\"c\":{\"1\":\"O\",\"2\":\"X\",\"3\":\"O\"}}},\"GameOver\":true}",
			},
			wantErr:  false,
			Name:     "Test",
			UserSide: "X",
		},
		{
			name:      "Сессии нет → возвращается ошибка",
			ctx:       context.Background(),
			sessionID: "333",
			fakeData: map[string]string{
				"game-session:12345": "{\"Name\":\"Test\",\"UserSide\":\"X\",\"BotSide\":\"O\",\"Builder\":{\"Unused\":0,\"Rows\":[\"a\",\"b\",\"c\"],\"Cols\":[1,2,3],\"Sides\":[\"X\",\"O\"],\"Board\":{\"a1\":1,\"a2\":2,\"a3\":1,\"b1\":2,\"b2\":1,\"b3\":1,\"c1\":2,\"c2\":1,\"c3\":2},\"BoardTemplate\":{\"a\":{\"1\":\"X\",\"2\":\"O\",\"3\":\"X\"},\"b\":{\"1\":\"O\",\"2\":\"X\",\"3\":\"X\"},\"c\":{\"1\":\"O\",\"2\":\"X\",\"3\":\"O\"}}},\"GameOver\":true}",
			},
			wantErr:     true,
			wantErrText: "session not found",
			Name:        "Test",
			UserSide:    "X",
		},
		{
			name:      "В Redis(fake) лежит некорректный JSON → возвращается ошибка.",
			ctx:       context.Background(),
			sessionID: "12345",
			fakeData: map[string]string{
				"game-session:12345": "{dasd}",
			},
			wantErr:     true,
			wantErrText: "not valid session",
			Name:        "Test",
			UserSide:    "X",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.ctx
			sessionId := tt.sessionID
			data := tt.fakeData

			service := NewGameService(&FakeStore{
				data: data,
			})

			gotSession, err := service.GetSession(ctx, sessionId)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if err != nil {
				if tt.wantErr && tt.wantErrText != err.Error() {
					t.Errorf("got = %v, want %v", err.Error(), tt.wantErrText)
				}
				return
			}

			if gotSession == nil {
				t.Fatal("expected session, got nil")
			}

			if gotSession.Name != "Test" {
				t.Errorf("got name %q, want %q", gotSession.Name, "Test")
			}
			if gotSession.UserSide != "X" {
				t.Errorf("got side %q, want %q", gotSession.UserSide, "X")
			}
		})
	}
}
