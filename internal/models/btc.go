package models

import (
	"encoding/json"
	"time"
)

type BTC struct {
	ID              string          `json:"id"  db:"id"`
	Value           float64         `json:"value" db:"value"`
	Latest          bool            `json:"latest" db:"latest"`
	CreatedAt       *time.Time      `json:"created_at" db:"created_at"`
	CurrenciesToBTC json.RawMessage `json:"currencies_to_btc" db:"currencies_to_btc"`
}
