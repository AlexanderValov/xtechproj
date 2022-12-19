package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	DB struct {
		URL string `envconfig:"DATABASE_URL" default:"postgres://postgres:strongPassword1@db:5432/postgres?sslmode=disable"`
	}
	PORT string `envconfig:"PORT" default:"8000"`
	URLs struct {
		BTCUSDT string `envconfig:"GET_BTCUSDT" default:"https://api.kucoin.com/api/v1/market/stats?symbol=BTC-USDT"`
		Fiat    string `envconfig:"GET_FIAT" default:"http://www.cbr.ru/scripts/XML_daily.asp"`
	}
}

func New() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
