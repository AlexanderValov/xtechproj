package services

import (
	"XTechProject/internal/models"
	"encoding/json"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func calculateBTCToFiat(currencies []models.Currency, btcInRub float64) (map[string]float64, error) {
	btcToFiat := make(map[string]float64, 34)
	for _, c := range currencies {
		btcToFiat[c.CharCode] = btcInRub / c.Val * float64(c.Nominal)
	}
	btcToFiat["RUB"] = btcInRub
	return btcToFiat, nil
}

func getResponse(link string) (*http.Response, error) {
	// NOTE: need to close resp
	resp, err := http.Get(link)
	if err != nil {
		return nil, fmt.Errorf("http.Get() err: %w", err)
	}
	// Success is indicated with 2xx status codes:
	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		return nil, fmt.Errorf("http.Get() status code: %d", resp.StatusCode)
	}
	return resp, nil
}

func serializeFiatCurrenciesData(val []Valute) ([]byte, float64, error) {
	if len(val) == 0 {
		return nil, 0, ErrEmptyValuteSlice
	}
	var (
		mu     = &sync.Mutex{}
		errs   = &errgroup.Group{}
		usdrub float64
	)
	cur := make([]models.Currency, 0, len(val))
	for _, v := range val {
		func(valute Valute) {
			errs.Go(func() error {
				nominal, err := strconv.Atoi(valute.Nominal)
				if err != nil {
					return err
				}
				value, err := strconv.ParseFloat(strings.Replace(valute.Value, ",", ".", 1), 64)
				if err != nil {
					return err
				}
				if valute.CharCode == models.CharCodeUSD {
					usdrub = value
				}
				mu.Lock()
				cur = append(cur, models.Currency{
					ID:       valute.ID,
					Name:     valute.Name,
					Nominal:  nominal,
					CharCode: valute.CharCode,
					NumCode:  valute.NumCode,
					Val:      value,
				})
				mu.Unlock()
				return nil
			})
		}(v)
	}
	if err := errs.Wait(); err != nil {
		return nil, 0, err
	}
	if usdrub == 0 {
		return nil, 0, ErrUSDNotFound
	}
	bts, err := json.Marshal(cur)
	if err != nil {
		return nil, 0, err
	}
	return bts, usdrub, nil
}

func serializeOrderBy(orderBy string) (string, error) {
	if orderBy != "" {
		switch orderBy {
		case "value", "created_at", "latest":
			orderBy = "ORDER BY " + orderBy
		case "-value", "-created_at", "-latest":
			orderBy = "ORDER BY " + orderBy[1:] + " DESC"
		default:
			return "", ErrUnexpectedOrderBy
		}
	}
	return orderBy, nil
}

func unixTimeToTime(unixTime int64) *time.Time {
	tm := time.Unix(0, unixTime*int64(time.Millisecond))
	return &tm
}
