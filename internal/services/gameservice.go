package services

import (
	"bohemas/internal/app"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type GameService struct {
	db *redis.Client
}

func NewGameService(redisDB *redis.Client) *GameService {
	return &GameService{
		db: redisDB,
	}
}

func (gs *GameService) GetSession(ctx context.Context, sessionID string) (*app.GameSession, error) {
	// Реалировать чтение сессии из редис
	session, err := gs.db.Get(ctx, "game-session:"+sessionID).Result()
	if errors.Is(err, redis.Nil) {
		// Ключ не существует в Redis
		return nil, errors.New("session not found")
	}

	if err != nil {
		// Другая ошибка: Redis недоступен и т.д.
		return nil, err
	}

	if session == "" {
		// Ключ существует, но значение — пустая строка
		return nil, errors.New("empty session")
	}

	var gameSession app.GameSession

	err = json.Unmarshal([]byte(session), &gameSession)
	if err != nil {
		return nil, err
	}

	fmt.Printf("gameSession: %v\n", gameSession.Name)

	return &gameSession, nil
}

func (gs *GameService) PlayStart(ctx context.Context, login, side string) (app.StartResult, error) {
	if strings.TrimSpace(login) == "" ||
		(side != "X" && side != "O") {
		err := errors.New("invalid login")
		return app.StartResult{}, err
	}

	botSide := "O"
	if side == "O" {
		botSide = "X"
	}

	id, err := app.NewSessionID()
	if err != nil {
		err = errors.New(fmt.Sprint("Error creating new session ID:", err))
		return app.StartResult{}, err
	}

	game := &app.GameSession{
		Name:     strings.TrimSpace(login),
		UserSide: side,
		BotSide:  botSide,
		Builder:  app.NewBoardBuilder(),
	}

	jsonSession, err := json.Marshal(game)
	if err != nil {
		err = errors.New(fmt.Sprint("Ошибка сериализации json сессии:", err))
		return app.StartResult{}, err
	}

	start := app.StartResult{
		SessionId: id,
		Game:      game,
	}

	// написать сохранение redis сессии вместо мутекса (с пробросом контекста)
	if err := gs.db.Set(ctx, "game-session:"+start.SessionId, jsonSession, 3000*time.Second).Err(); err != nil {
		return start, err
	}

	return start, nil
}

func (gs *GameService) PlayMove(ctx context.Context, sessionId string, position string) (app.MoveResult, error) {
	move := app.MoveResult{
		BotMove: "",
		Winner:  "",
		IsDraw:  false,
	}

	// Реалировать чтение сессии из редис
	session, err := gs.db.Get(ctx, "game-session:"+sessionId).Result()
	if errors.Is(err, redis.Nil) {
		// Ключ не существует в Redis
		return move, errors.New("session not found")
	}

	if err != nil {
		// Другая ошибка: Redis недоступен и т.д.
		return app.MoveResult{}, err
	}

	if session == "" {
		// Ключ существует, но значение — пустая строка
		return app.MoveResult{}, err
	}
	var game *app.GameSession
	err = json.Unmarshal([]byte(session), &game)
	if err != nil {
		return app.MoveResult{}, err
	}

	if game.GameOver {
		err := errors.New("game already finished")
		return app.MoveResult{}, err
	}

	value, exists := game.Builder.Board[position]
	if !exists {
		err := errors.New("Invalid cell")
		return app.MoveResult{}, err
	}
	if value != game.Builder.Unused {
		err := errors.New("Cell already occupied")
		return app.MoveResult{}, err
	}

	game.Builder.Board[position] = app.Player1 // выбор юзера
	chars := strings.Split(position, "")
	rowNumber, _ := strconv.Atoi(chars[1])
	game.Builder.BoardTemplate[chars[0]][rowNumber] = game.UserSide

	// Группировка инфы по сторонам пользователей
	players := []string{
		game.UserSide,
		game.BotSide,
	}

	if win, winner := app.IsWinner(game.Builder, players); win {
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
		move.BotMove = app.BotMove(game.Builder, game.BotSide, game.UserSide)

		if move.BotMove != "" {
			game.Builder.Board[move.BotMove] = app.Player2

			// Формирование стороны и выбора BOT
			chars2 := strings.Split(move.BotMove, "")
			rowNumber2, _ := strconv.Atoi(chars2[1])
			game.Builder.BoardTemplate[chars2[0]][rowNumber2] = game.BotSide

			if win, winner := app.IsWinner(game.Builder, players); win {
				move.Winner = winner
			}
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

	jsonSession, err := json.Marshal(game)
	if err != nil {
		err = errors.New(fmt.Sprint("Ошибка сериализации json сессии:", err))
		return move, err
	}

	if err := gs.db.Set(ctx, "game-session:"+sessionId, jsonSession, 3000*time.Second).Err(); err != nil {
		return move, err
	}

	return move, nil
}

func (gs *GameService) PlayReset(ctx context.Context, sessionId string) error {
	// Реалировать чтение сессии из редис
	session, err := gs.db.Get(ctx, "game-session:"+sessionId).Result()
	if errors.Is(err, redis.Nil) {
		// Ключ не существует в Redis
		return err
	}

	if err != nil {
		// Другая ошибка: Redis недоступен и т.д.
		return err
	}

	if session == "" {
		// Ключ существует, но значение — пустая строка
		return errors.New("session not found")
	}
	var game *app.GameSession
	err = json.Unmarshal([]byte(session), &game)
	if err != nil {
		return err
	}

	game.Builder = app.NewBoardBuilder()
	game.GameOver = false

	jsonSession, err := json.Marshal(game)
	if err != nil {
		err = errors.New(fmt.Sprint("Ошибка сериализации json сессии:", err))
		return err
	}

	if err := gs.db.Set(ctx, "game-session:"+sessionId, jsonSession, 3000*time.Second).Err(); err != nil {
		return err
	}

	return nil
}
