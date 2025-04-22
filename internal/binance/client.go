package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Tick struct {
	Symbol    string
	Price     float64
	Quantity  float64
	Timestamp time.Time
}

type Client struct {
	wssURL  string
	symbols []string
}

func NewClient(wssURL string, symbols []string) *Client {
	return &Client{
		wssURL:  wssURL,
		symbols: symbols,
	}
}

func (c *Client) Connect(ctx context.Context, tickChan chan<- Tick) {
	var wg sync.WaitGroup
	for _, sym := range c.symbols {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()
			c.connectSymbol(ctx, symbol, tickChan)
		}(sym)
	}
	wg.Wait()
	close(tickChan)
}

func (c *Client) connectSymbol(ctx context.Context, symbol string, tickChan chan<- Tick) {
	url := fmt.Sprintf("%s/%s@aggTrade", c.wssURL, strings.ToLower(symbol))

	log.Printf("Connecting to %s", url)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("dial %s error: %v", symbol, err)
			time.Sleep(time.Second)
			continue
		}

		func() {
			defer conn.Close()
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Printf("read %s error: %v", symbol, err)
					return
				}

				var t struct {
					Symbol    string `json:"s"`
					Price     string `json:"p"`
					Quantity  string `json:"q"`
					Timestamp int64  `json:"T"`
				}
				if err := json.Unmarshal(message, &t); err != nil {
					log.Printf("unmarshal %s error: %v", symbol, err)
					continue
				}

				price, _ := strconv.ParseFloat(t.Price, 64)
				qty, _ := strconv.ParseFloat(t.Quantity, 64)
				tickChan <- Tick{
					Symbol:    t.Symbol,
					Price:     price,
					Quantity:  qty,
					Timestamp: time.Unix(0, t.Timestamp*int64(time.Millisecond)),
				}
			}
		}()
	}
}
