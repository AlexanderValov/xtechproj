package models

import (
	"encoding/json"
	"time"
)

type BTC struct {
	ID        int             `json:"id"  db:"id"`
	Value     float64         `json:"value" db:"value"`
	Latest    bool            `json:"latest" db:"latest"`
	CreatedAt *time.Time      `json:"created_at" db:"created_at"`
	BTCToFiat json.RawMessage `json:"btc_to_fiat" db:"btc_to_fiat"`
}
