package aggregator

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/shubie/trading/internal/binance"
)

type Candle struct {
	Symbol    string    `db:"symbol"`
	Open      float64   `db:"open"`
	High      float64   `db:"high"`
	Low       float64   `db:"low"`
	Close     float64   `db:"close"`
	Volume    float64   `db:"volume"`
	StartTime time.Time `db:"start_time"`
	EndTime   time.Time `db:"end_time"`
	Finalized bool      `db:"-"`
}

type Aggregator struct {
	mu           sync.RWMutex
	candles      map[string]*Candle
	lastTickTime time.Time
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		candles: make(map[string]*Candle),
	}
}

func (a *Aggregator) Run(ctx context.Context, tickChan <-chan binance.Tick, candleChan chan<- Candle) {
	defer close(candleChan)
	log.Println("Aggregator service started")
	finalizeTicker := time.NewTicker(1 * time.Second)
	defer finalizeTicker.Stop()

	for {
		select {
		case tick, ok := <-tickChan:
			if !ok {
				a.finalizeAll(candleChan)
				log.Println("Tick channel closed, shutting down aggregator")
				return
			}
			log.Printf("Aggregator received tick: %+v\n", tick) // ← ADD THIS
			a.processTick(tick)

		case <-finalizeTicker.C:
			a.finalizeExpired(candleChan)

		case <-ctx.Done():
			a.finalizeAll(candleChan)
			log.Println("Context cancelled, shutting down aggregator")
			return
		}
	}
}

func (a *Aggregator) processTick(tick binance.Tick) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.lastTickTime = tick.Timestamp

	startTime := tick.Timestamp.Truncate(time.Minute)
	key := tick.Symbol + startTime.String()

	candle, exists := a.candles[key]
	if !exists {
		candle = &Candle{
			Symbol:    tick.Symbol,
			Open:      tick.Price,
			High:      tick.Price,
			Low:       tick.Price,
			Close:     tick.Price,
			Volume:    tick.Quantity,
			StartTime: startTime,
			EndTime:   startTime.Add(time.Minute),
		}
		a.candles[key] = candle
		return
	}

	if tick.Price > candle.High {
		candle.High = tick.Price
	}
	if tick.Price < candle.Low {
		candle.Low = tick.Price
	}
	candle.Close = tick.Price
	candle.Volume += tick.Quantity
}

func (a *Aggregator) finalizeExpired(candleChan chan<- Candle) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	for key, candle := range a.candles {
		if now.After(candle.EndTime) {
			candle.Finalized = true
			candleChan <- *candle
			delete(a.candles, key)
			log.Printf("Finalized candle for %s: %s-%s",
				candle.Symbol,
				candle.StartTime.Format(time.RFC3339),
				candle.EndTime.Format(time.RFC3339))
		}
	}
}

func (a *Aggregator) finalizeAll(candleChan chan<- Candle) {
	a.mu.Lock()
	defer a.mu.Unlock()
	log.Println("Finalizing all remaining candles")

	for key, candle := range a.candles {
		candle.Finalized = true
		candleChan <- *candle
		delete(a.candles, key)
	}
}

func (a *Aggregator) GetCurrentCandle(symbol string) *Candle {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, candle := range a.candles {
		if candle.Symbol == symbol && !candle.Finalized {
			return candle
		}
	}
	return nil
}

func (a *Aggregator) GetLastDataTime() time.Time {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastTickTime
}

func (a *Aggregator) SetLastDataTimeForTesting(t time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastTickTime = t
}
