package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/shubie/trading/internal/aggregator"
	"github.com/shubie/trading/internal/binance"
	"github.com/shubie/trading/internal/config"
	"github.com/shubie/trading/internal/grpcserver"
	"github.com/shubie/trading/internal/health"
	"github.com/shubie/trading/internal/storage"
)

func main() {
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatal("Config error:", err)
	}

	store := storage.NewPostgresStorage(cfg.Storage.Postgres.DSN)
	agg := aggregator.NewAggregator()
	grpcServer := grpcserver.NewServer(cfg.GRPC.Port, agg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tickChan := make(chan binance.Tick, cfg.Buffers.TickChan)
	candleChan := make(chan aggregator.Candle, cfg.Buffers.CandleChan)

	var wg sync.WaitGroup

	binanceClient := binance.NewClient(cfg.Binance.WSSURL, cfg.Binance.Symbols)
	wg.Add(1)
	go func() {
		defer wg.Done()
		binanceClient.Connect(ctx, tickChan)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		agg.Run(ctx, tickChan, candleChan)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		grpcServer.Start()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		store.StartPersisting(ctx, candleChan)
	}()

	healthHandler := health.NewHandler(agg, cfg.Health.DataTimeout)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("HTTP health server starting on port %d", cfg.Health.Port)
		if err := http.ListenAndServe(
			fmt.Sprintf(":%d", cfg.Health.Port),
			healthHandler,
		); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("\nInitiating shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	cancel()
	grpcServer.Stop()
	store.Close()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Graceful shutdown completed")
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout exceeded")
	}
}
