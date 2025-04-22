package binance_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shubie/trading/internal/binance"

	"github.com/gorilla/websocket"
)

func TestClient_Connect(t *testing.T) {
	// Setup WebSocket mock server
	symbol := "BTCUSDT"
	upgrader := websocket.Upgrader{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, strings.ToLower(symbol)+"@aggTrade") {
			http.Error(w, "invalid endpoint", http.StatusNotFound)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade websocket: %v", err)
		}
		defer conn.Close()

		// Send a mock message
		msg := map[string]interface{}{
			"s": symbol,
			"p": "45000.00",
			"q": "0.001",
			"T": time.Now().UnixMilli(),
		}
		data, _ := json.Marshal(msg)
		time.Sleep(100 * time.Millisecond) // small delay to allow client to start reading
		conn.WriteMessage(websocket.TextMessage, data)
		time.Sleep(200 * time.Millisecond)
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	client := binance.NewClient(wsURL, []string{symbol})
	tickChan := make(chan binance.Tick)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go client.Connect(ctx, tickChan)

	select {
	case tick := <-tickChan:
		if tick.Symbol != symbol {
			t.Errorf("expected symbol %s, got %s", symbol, tick.Symbol)
		}
		if tick.Price != 45000.00 {
			t.Errorf("expected price 45000.00, got %f", tick.Price)
		}
		if tick.Quantity != 0.001 {
			t.Errorf("expected quantity 0.001, got %f", tick.Quantity)
		}
	case <-time.After(time.Second * 2):
		t.Fatal("Timeout waiting for tick data")
	}
}
