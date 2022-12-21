package models

import (
	"encoding/json"
	"time"
)

type BTC struct {
	ID        int             `json:"id"  db:"id"`
	InUSDT    float64         `json:"in_usdt" db:"in_usdt"`
	InRub     float64         `json:"in_rub" db:"in_rub"`
	Latest    bool            `json:"latest" db:"latest"`
	CreatedAt *time.Time      `json:"created_at" db:"created_at"`
	BTCToFiat json.RawMessage `json:"btc_to_fiat" db:"btc_to_fiat"`
}
