package services

import (
	"XTechProject/cmd/config"
	"XTechProject/internal/models"
	"XTechProject/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

var (
	ErrUSDNotFound             = errors.New("USD not found")
	ErrEmptyValuteSlice        = errors.New("empty valutes slice")
	ErrUnexpectedOrderBy       = errors.New("unexpected order_by")
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
		GetBTCToFiat(btc *models.BTC) (*map[string]float64, error)

		GetLastFiat() (*models.Fiat, error)
		GetFiatHistory(limit, offset int, orderBy string) ([]models.Fiat, error)
		CheckLastDateUpdatingFiatCurrencies() error
	}
)

func NewManagementService(db repository.Repositorier, cfg *config.Config) *ManagementService {
	svc := &ManagementService{db: db, cfg: cfg}
	return svc
}

func (svc *ManagementService) RunWorkers() {
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

func (svc *ManagementService) UpdateBTCInDB(unixTime int64, lastValue string) {
	if err := svc.db.UpdateLastRecordForBTC(); err != nil {
		log.Printf("BTCWorker: error in UpdateLastRecordForBTC, err %s\n", err)
	}
	inUSDT, err := strconv.ParseFloat(lastValue, 64)
	if err != nil {
		log.Printf("BTCWorker: error in ParseFloat(lastValue, 64), err %s\n", err)
	}
	btc := &models.BTC{
		InUSDT:    inUSDT,
		CreatedAt: unixTimeToTime(unixTime),
		Latest:    true,
	}
	if err = svc.db.CreateBTCRecord(btc); err != nil {
		log.Printf("BTCWorker: error in CreateBTCRecord, err %s\n", err)
	}
	log.Println("BTC updated in db")
	if err := svc.UpdateBTCToFiatInDB(btc); err != nil {
		log.Printf("BTCWorker: error in UpdateBTCToFiatInDB, err %s\n", err)
	}
}

func (svc *ManagementService) UpdateBTCToFiatInDB(btc *models.BTC) error {
	btcToFiat, err := svc.GetBTCToFiat(btc)
	if err != nil {
		return fmt.Errorf("error in GetBTCToFiat(btc), err: %w", err)
	}
	btc.BTCToFiat, err = json.Marshal(btcToFiat)
	if err != nil {
		return fmt.Errorf("error in json.Marshal(btcToFiat), err: %w", err)
	}
	if err = svc.db.UpdateFiatForLastBTC(btc); err != nil {
		return fmt.Errorf("error in UpdateFiatForLastBTC(btc), err: %w", err)
	}
	log.Println("BTC/Fiat updated in db")
	return nil
}

func (svc *ManagementService) GetBTCToFiat(btc *models.BTC) (*map[string]float64, error) {
	lastFiat, err := svc.db.GetLastFiat()
	if err != nil {
		return nil, fmt.Errorf("error in GetLastFiat: %w", err)
	}
	var currencies []models.Currency
	if err := json.Unmarshal(lastFiat.Currencies, &currencies); err != nil {
		return nil, fmt.Errorf("error in json.Unmarshal: %w", err)
	}
	btc.InRub = btc.InUSDT * lastFiat.USDRUB
	btcToFiat, err := calculateBTCToFiat(currencies, btc.InRub)
	if err != nil {
		return nil, err
	}
	return &btcToFiat, nil
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
