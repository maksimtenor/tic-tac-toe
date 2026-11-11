package app

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

var sessions = make(map[string]*GameSession)
var sessionsMu sync.Mutex

type GameSession struct {
	Name     string
	UserSide string
	BotSide  string
	Builder  *BoardBuilder
	GameOver bool
	Mu       sync.Mutex
}

func newSessionID() (string, error) {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
