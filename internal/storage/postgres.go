package storage

import (
	"context"
	_ "database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"embed"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/shubie/trading/internal/aggregator"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type PostgresStorage struct {
	db *sqlx.DB
}

func NewPostgresStorage(dsn string) *PostgresStorage {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		panic(fmt.Sprintf("DB connection error: %v", err))
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	runMigrations(dsn)

	return &PostgresStorage{db: db}
}

func runMigrations(dsn string) {
	driver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		log.Fatal("Failed to create migration driver:", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", driver, dsn)
	if err != nil {
		log.Fatal("Migration setup failed:", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("Migration failed:", err)
	}
}

func (s *PostgresStorage) StartPersisting(ctx context.Context, candleChan <-chan aggregator.Candle) {
	go func() {
		log.Println("Persistence worker started")
		batch := make([]aggregator.Candle, 0, 100)
		ticker := time.NewTicker(1 * time.Second)
		defer func() {
			ticker.Stop()
			if len(batch) > 0 {
				s.persistBatch(batch)
				log.Printf("Persisted final batch of %d candles\n", len(batch))
			}
			log.Println("Persistence worker stopped")
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case candle, ok := <-candleChan:
				if !ok {
					return
				}
				batch = append(batch, candle)
				if len(batch) >= 100 {
					s.persistBatch(batch)
					batch = batch[:0]
				}

			case <-ticker.C:
				if len(batch) > 0 {
					s.persistBatch(batch)
					batch = batch[:0]
				}
			}
		}
	}()
}

func (s *PostgresStorage) persistBatch(candles []aggregator.Candle) {
	if len(candles) == 0 {
		return
	}

	query := `
        INSERT INTO candlesticks 
        (symbol, open, high, low, close, volume, start_time, end_time)
        VALUES (:symbol, :open, :high, :low, :close, :volume, :start_time, :end_time)
        ON CONFLICT (symbol, start_time) DO NOTHING`

	_, err := s.db.NamedExec(query, candles)
	if err != nil {
		log.Printf("Persistence error: %v", err)
		return
	}

	log.Printf("Persisted %d candles", len(candles))
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}
