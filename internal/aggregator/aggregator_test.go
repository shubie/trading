package aggregator_test

import (
	"context"
	"testing"
	"time"

	"github.com/shubie/trading/internal/aggregator"
	"github.com/shubie/trading/internal/binance"
)

func TestAggregator_Run(t *testing.T) {
	tickChan := make(chan binance.Tick)
	candleChan := make(chan aggregator.Candle)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	agg := aggregator.NewAggregator()
	go agg.Run(ctx, tickChan, candleChan)

	// I am sending a test tick to the tickChan channel
	now := time.Now()
	tick := binance.Tick{
		Symbol:    "BTCUSDT",
		Price:     10000.0,
		Quantity:  0.5,
		Timestamp: now,
	}
	tickChan <- tick

	time.Sleep(2 * time.Second)
	cancel()

	var finalizedCandle aggregator.Candle
	select {
	case finalizedCandle = <-candleChan:
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for finalized candle")
	}

	if finalizedCandle.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %s", finalizedCandle.Symbol)
	}
	if finalizedCandle.Open != 10000.0 || finalizedCandle.Close != 10000.0 {
		t.Errorf("Expected open/close to be 10000.0, got %f/%f", finalizedCandle.Open, finalizedCandle.Close)
	}
	if !finalizedCandle.Finalized {
		t.Errorf("Expected candle to be finalized")
	}
}
