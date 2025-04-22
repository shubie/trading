package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Binance struct {
		WSSURL  string   `mapstructure:"wss_url"`
		Symbols []string `mapstructure:"symbols"`
	}
	GRPC struct {
		Port int `mapstructure:"port"`
	}
	Storage struct {
		Postgres struct {
			DSN string `mapstructure:"dsn"`
		}
	}
	Buffers struct {
		TickChan   int `mapstructure:"tick_chan"`
		CandleChan int `mapstructure:"candle_chan"`
	}
	Health struct {
		DataTimeout time.Duration `mapstructure:"data_timeout"`
		Port        int           `mapstructure:"port"`
	}
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	viper.SetDefault("health.data_timeout", 5*time.Minute)
	viper.SetDefault("buffers.tick_chan", 1000)
	viper.SetDefault("buffers.candle_chan", 500)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
