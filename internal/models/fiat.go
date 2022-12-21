package models

import (
	"encoding/json"
	"time"
)

const CharCodeUSD = "USD"

type (
	Fiat struct {
		ID         int             `json:"id"  db:"id"`
		Latest     bool            `json:"latest" db:"latest"`
		CreatedAt  *time.Time      `json:"created_at" db:"created_at"`
		USDRUB     float64         `json:"usd_rub" db:"usd_rub"`
		Currencies json.RawMessage `json:"currencies" db:"currencies"`
	}
	Currency struct {
		ID       string  `json:"id"  db:"id"`
		Nominal  int     `json:"nominal" db:"nominal"`
		Name     string  `json:"name" db:"name"`
		Val      float64 `json:"value" db:"value"`
		CharCode string  `json:"char_code" db:"char_code"`
		NumCode  string  `json:"num_code" db:"num_code"`
	}
)
