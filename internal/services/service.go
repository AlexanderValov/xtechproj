package services

import (
	"XTechProject/cmd/config"
	"XTechProject/internal/models"
	"XTechProject/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrUSDNotFound             = errors.New("USD not found")
	ErrEmptyValuteSlice        = errors.New("empty valutes slice")
	ErrUnexpectedType          = errors.New("unexpected type for order_by")
	ErrAlreadyUpdatedFiatToday = errors.New("fiat currencies were already updated today")
)

type (
	ManagementService struct {
		db  repository.Repositorier
		cfg *config.Config
	}
	Servicer interface {
		GetLastBTC() (*models.BTC, error)
		GetAllBTC(limit, offset int, orderBy string) ([]models.BTC, error)
		UpdateBTCInDB(unixTime int64, lastValue string) error
		GetBTCToFiat(btc *models.BTC) (*map[string]float64, error)

		GetLastFiat() (*models.Fiat, error)
		GetFiatHistory(limit, offset int, orderBy string) ([]models.Fiat, error)
		CheckLastDateUpdatingFiatCurrencies() error
	}
)

func NewManagementService(db repository.Repositorier, cfg *config.Config) *ManagementService {
	svc := &ManagementService{db: db, cfg: cfg}
	// run workers
	go svc.runWorkers()
	return svc
}

func (svc *ManagementService) runWorkers() {
	// first starting after running server
	go svc.BTCWorker()
	// fiat will not created if it was already created today
	go svc.FiatWorker()
	// tickers will trigger workers
	tickerForBTC := time.NewTicker(time.Second * 10).C
	tickerForFiat := time.NewTicker(time.Hour * 24).C
	for {
		select {
		case <-tickerForBTC:
			go svc.BTCWorker()
		case <-tickerForFiat:
			go svc.FiatWorker()
		}
	}
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
		usdrub float64
	)
	errCh := make(chan error)
	defer close(errCh)
	cur := make([]models.Currency, 0, len(val))
	for _, v := range val {
		go func(valute Valute) {
			nominal, err := strconv.Atoi(valute.Nominal)
			if err != nil {
				errCh <- err
				return
			}
			value, err := strconv.ParseFloat(strings.Replace(valute.Value, ",", ".", 1), 64)
			if err != nil {
				errCh <- err
				return
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
			if len(cur) == len(val) {
				errCh <- nil
			}
		}(v)
	}
	if err := <-errCh; err != nil {
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
			return "", ErrUnexpectedType
		}
	}
	return orderBy, nil
}

func (svc *ManagementService) UpdateBTCInDB(unixTime int64, lastValue string) error {
	if err := svc.db.UpdateLastRecordForBTC(); err != nil {
		log.Printf("BTCWorker: error in UpdateLastRecordForBTC, err %s\n", err)
		return err
	}
	tm := time.Unix(0, unixTime*int64(time.Millisecond))
	value, err := strconv.ParseFloat(lastValue, 64)
	if err != nil {
		log.Printf("BTCWorker: error in ParseFloat, err %s\n", err)
		return err
	}
	btc := &models.BTC{
		Value:     value,
		CreatedAt: &tm,
		Latest:    true,
	}
	btcToFiat, err := svc.GetBTCToFiat(btc)
	if err != nil {
		log.Printf("BTCWorker: error in GetBTCToFiat, err %s\n", err)
		return err
	}
	btc.CurrenciesToBTC, err = json.Marshal(btcToFiat)
	if err != nil {
		log.Printf("BTCWorker: error in json.Marshal, err %s\n", err)
		return err
	}
	if err = svc.db.CreateBTCRecord(btc); err != nil {
		log.Printf("BTCWorker: error in CreateBTCRecord, err %s\n", err)
		return err
	}
	log.Println("BTC/USDT and BTC/Fiat updated in db")
	return nil
}

func (svc *ManagementService) CheckLastDateUpdatingFiatCurrencies() error {
	date, err := svc.db.GetLastDateForFiat()
	if err != nil {
		return fmt.Errorf("error in GetLastDateForFiaty, err: %s\n", err.Error())
	}
	if date != nil {
		// if we already have data today -> skip
		d := date.Format(time.RFC3339[:10])
		t := time.Now().Format(time.RFC3339[:10])
		if d == t {
			return ErrAlreadyUpdatedFiatToday
		}
	}
	return nil
}

func (svc *ManagementService) GetLastBTC() (*models.BTC, error) {
	model, err := svc.db.GetLastBTC()
	if err != nil {
		return nil, fmt.Errorf("error in GetLastBTC: %w", err)
	}
	return model, nil
}

func (svc *ManagementService) GetAllBTC(limit, offset int, orderBy string) ([]models.BTC, error) {
	orderBy, err := serializeOrderBy(orderBy)
	if err != nil {
		return nil, fmt.Errorf("error in serializeOrderBy: %w", err)
	}
	modelsData, err := svc.db.GetAllBTC(limit, offset, orderBy)
	if err != nil {
		return nil, fmt.Errorf("error to get all btcusdt data, err: %w", err)
	}
	return modelsData, nil
}

func (svc *ManagementService) GetLastFiat() (*models.Fiat, error) {
	model, err := svc.db.GetLastFiat()
	if err != nil {
		return nil, fmt.Errorf("error in GetLastFiat: %w", err)
	}
	return model, nil
}

func (svc *ManagementService) GetFiatHistory(limit, offset int, orderBy string) ([]models.Fiat, error) {
	orderBy, err := serializeOrderBy(orderBy)
	if err != nil {
		return nil, fmt.Errorf("error in serializeOrderBy: %w", err)
	}
	modelsData, err := svc.db.GetAllFiat(limit, offset, orderBy)
	if err != nil {
		return nil, fmt.Errorf("error in GetAllFiat: %w", err)
	}
	return modelsData, nil
}

func (svc *ManagementService) GetBTCToFiat(btc *models.BTC) (*map[string]float64, error) {
	lastFiat, err := svc.db.GetLastFiat()
	if err != nil {
		return nil, fmt.Errorf("error in GetLastFiat: %w", err)
	}
	btcToFiat := make(map[string]float64, 34)
	var currencies []models.Currency
	err = json.Unmarshal(lastFiat.Currencies, &currencies)
	if err != nil {
		return nil, fmt.Errorf("error in json.Unmarshal: %w", err)
	}
	for _, c := range currencies {
		btcToFiat[c.CharCode] = btc.Value * lastFiat.USDRUB / c.Val * float64(c.Nominal)
	}
	return &btcToFiat, nil
}
