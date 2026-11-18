package rest

import (
	"bohemas/internal/services"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestMoveHandler(t *testing.T) {
	gameService := services.NewGameService(&FakeStore{
		make(map[string]string),
	})

	body := strings.NewReader("{\"login\":\"Test\",\"side\":\"X\"}")

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/start",
		body,
	)

	rec := httptest.NewRecorder()

	router := NewRouter(gameService)
	router.ServeHTTP(rec, req)

	resp := rec.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got %d; want %d", resp.StatusCode, http.StatusOK)
	}

	var cookieData = http.Cookie{}
	for _, cookie := range resp.Cookies() {
		if cookie.Name != "ttt_session" {
			t.Fatal("cookie name is wrong")
		}
		cookieData.Name = cookie.Name
		if !cookie.HttpOnly {
			t.Error("cookie must be HttpOnly")
		}

		if cookie.Value == "" {
			t.Error("session cookie is empty")
		} else {
			cookieData.Value = cookie.Value
		}
	}

	body = strings.NewReader("{\"position\":\"a1\"}")

	req = httptest.NewRequest(
		http.MethodPost,
		"/api/move",
		body,
	)
	req.AddCookie(&cookieData)

	rec = httptest.NewRecorder()

	router = NewRouter(gameService)
	router.ServeHTTP(rec, req)

	resp = rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got %d; want %d", resp.StatusCode, http.StatusOK)
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name != "ttt_session" {
			t.Fatal("cookie name is wrong")
		}
		cookieData.Name = cookie.Name
		if !cookie.HttpOnly {
			t.Error("cookie must be HttpOnly")
		}

		if cookie.Value == "" {
			t.Error("session cookie is empty")
		} else {
			cookieData.Value = cookie.Value
		}
	}

	req = httptest.NewRequest(
		http.MethodGet,
		"/api/session",
		strings.NewReader(""),
	)
	req.AddCookie(&cookieData)

	rec = httptest.NewRecorder()

	router = NewRouter(gameService)
	router.ServeHTTP(rec, req)

	resp = rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		var result struct {
			UserSide  string                    `json:"userSide"`
			BotSide   string                    `json:"botSide"`
			SessionId string                    `json:"sessionId"`
			Name      string                    `json:"name"`
			Board     map[string]map[int]string `json:"board"`
			GameOver  bool                      `json:"gameOver"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		if result.Board["a"][1] != "X" {
			t.Errorf("got %s; want %s", result.Board["a"][1], "X")
		}
	} else {
		t.Fatalf("got %d; want %d", resp.StatusCode, http.StatusOK)
	}

}

func TestResetHandler(t *testing.T) {
	gameService := services.NewGameService(&FakeStore{
		make(map[string]string),
	})

	body := strings.NewReader("{\"login\":\"Test\",\"side\":\"X\"}")

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/start",
		body,
	)

	rec := httptest.NewRecorder()

	router := NewRouter(gameService)
	router.ServeHTTP(rec, req)

	resp := rec.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got %d; want %d", resp.StatusCode, http.StatusOK)
	}

	var cookieData = http.Cookie{}
	for _, cookie := range resp.Cookies() {
		if cookie.Name != "ttt_session" {
			t.Fatal("cookie name is wrong")
		}
		cookieData.Name = cookie.Name
		if !cookie.HttpOnly {
			t.Error("cookie must be HttpOnly")
		}

		if cookie.Value == "" {
			t.Error("session cookie is empty")
		} else {
			cookieData.Value = cookie.Value
		}
	}

	body = strings.NewReader("{\"position\":\"a1\"}")

	req = httptest.NewRequest(
		http.MethodPost,
		"/api/move",
		body,
	)
	req.AddCookie(&cookieData)

	rec = httptest.NewRecorder()

	router = NewRouter(gameService)
	router.ServeHTTP(rec, req)

	resp = rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got %d; want %d", resp.StatusCode, http.StatusOK)
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name != "ttt_session" {
			t.Fatal("cookie name is wrong")
		}
		cookieData.Name = cookie.Name
		if !cookie.HttpOnly {
			t.Error("cookie must be HttpOnly")
		}

		if cookie.Value == "" {
			t.Error("session cookie is empty")
		} else {
			cookieData.Value = cookie.Value
		}
	}

	body = strings.NewReader(cookieData.Value)

	req = httptest.NewRequest(
		http.MethodPost,
		"/api/reset",
		body,
	)
	req.AddCookie(&cookieData)

	rec = httptest.NewRecorder()

	router = NewRouter(gameService)
	router.ServeHTTP(rec, req)

	resp = rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got %d; want %d", resp.StatusCode, http.StatusOK)
	}

	req = httptest.NewRequest(
		http.MethodGet,
		"/api/session",
		strings.NewReader(""),
	)
	req.AddCookie(&cookieData)

	rec = httptest.NewRecorder()

	router = NewRouter(gameService)
	router.ServeHTTP(rec, req)

	resp = rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		var result struct {
			Board map[string]map[int]string `json:"board"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		if result.Board["a"][1] != "_" {
			t.Errorf("got %s; want %s", result.Board["a"][1], "X")
		}
	} else {
		t.Fatalf("got %d; want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestSessionHandler(t *testing.T) {
	gameService := services.NewGameService(&FakeStore{
		make(map[string]string),
	})

	body := strings.NewReader("{\"login\":\"Test\",\"side\":\"X\"}")

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/start",
		body,
	)

	rec := httptest.NewRecorder()

	router := NewRouter(gameService)
	router.ServeHTTP(rec, req)

	resp := rec.Result()

	if resp.StatusCode == http.StatusOK {
		var result struct {
			SessionID string `json:"sessionId"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}

		if result.SessionID == "" {
			t.Error("response sessionId is empty")
		}
	}
	var cookieData = http.Cookie{}
	for _, cookie := range resp.Cookies() {
		if cookie.Name != "ttt_session" {
			t.Fatal("cookie name is wrong")
		}
		cookieData.Name = cookie.Name
		if !cookie.HttpOnly {
			t.Error("cookie must be HttpOnly")
		}

		if cookie.Value == "" {
			t.Error("session cookie is empty")
		} else {
			cookieData.Value = cookie.Value
		}
	}

	body = strings.NewReader(cookieData.Value)

	req = httptest.NewRequest(
		http.MethodGet,
		"/api/session",
		body,
	)

	req.AddCookie(&cookieData)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	resp = rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result struct {
		SessionID string `json:"sessionId"`
		Name      string `json:"name"`
		UserSide  string `json:"userSide"`
		BotSide   string `json:"botSide"`
		GameOver  bool   `json:"gameOver"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if result.SessionID == "" {
		t.Error("response sessionId is empty")
	}

	if result.Name != "Test" {
		t.Error("response name is wrong")
	}
	if result.UserSide != "X" {
		t.Error("response userSide is wrong")
	}
	if result.BotSide != "O" {
		t.Error("response botSide is wrong")
	}
	if result.GameOver {
		t.Error("response gameOver is true")
	}
}

func TestStartHandler(t *testing.T) {
	test := []struct {
		name       string
		url        string
		waitStatus int
		fakeData   string
	}{
		{
			name:       "Корректный JSON",
			fakeData:   "{\"login\":\"Test\",\"side\":\"X\"}",
			waitStatus: http.StatusOK,
		},
		{
			name:       "Пустой login",
			fakeData:   "{\"login\":\"\",\"side\":\"X\"}",
			waitStatus: http.StatusBadRequest,
		},
		{
			name:       "Некорректный JSON",
			fakeData:   "{invalid json}",
			waitStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			gameService := services.NewGameService(&FakeStore{
				make(map[string]string),
			})

			body := strings.NewReader(tt.fakeData)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/start",
				body,
			)

			rec := httptest.NewRecorder()

			router := NewRouter(gameService)
			router.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				var result struct {
					SessionID string `json:"sessionId"`
				}

				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatal(err)
				}

				if result.SessionID == "" {
					t.Error("response sessionId is empty")
				}
			}

			found := false

			for _, cookie := range resp.Cookies() {
				if cookie.Name == "ttt_session" {
					found = true

					if !cookie.HttpOnly {
						t.Error("cookie must be HttpOnly")
					}

					if cookie.Value == "" {
						t.Error("session cookie is empty")
					}
				}
			}

			wantCookie := tt.waitStatus == http.StatusOK

			if found != wantCookie {
				t.Errorf("got cookie=%v, want cookie=%v", found, wantCookie)
			}

			if tt.waitStatus != resp.StatusCode {
				t.Errorf("Got: %d, want: %d", resp.StatusCode, tt.waitStatus)
			}
		})
	}
}
