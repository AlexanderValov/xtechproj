package services

import (
	"XTechProject/cmd/config"
	"XTechProject/internal/models"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/net/html/charset"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	repositorier interface {
		CreateBTCRecord(model *models.BTC) error
		UpdateLastRecordForBTC() error
		GetLastBTC() (*models.BTC, error)
		GetAllBTC(limit, offset int, orderBy string) ([]models.BTC, error)

		GetLastFiat() (*models.Fiat, error)
		GetAllFiat(limit, offset int, orderBy string) ([]models.Fiat, error)
		CreateFiatRecord(model *models.Fiat) error
		SetAllRecordsFiatLatestFalse() error
		GetLastDateForFiat() (*time.Time, error)
	}

	ManagementService struct {
		db  repositorier
		cfg *config.Config
	}
)

func NewManagementService(db repositorier, cfg *config.Config) *ManagementService {
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
	// tickers will push workers
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
	// NOTE: need to close resp -> resp.Body.Close()

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
	var (
		mu     = &sync.Mutex{}
		wg     sync.WaitGroup
		usdrub float64
	)
	cur := make([]models.Currency, 0, len(val))
	for _, v := range val {
		wg.Add(1)
		go func(valute Valute) {
			defer wg.Done()
			nominal, _ := strconv.Atoi(valute.Nominal)
			value, _ := strconv.ParseFloat(strings.Replace(valute.Value, ",", ".", 1), 64)
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
		}(v)
	}
	wg.Wait()
	bts, err := json.Marshal(cur)
	if err != nil {
		log.Println(err)
	}
	return bts, usdrub, nil
}

func serializeOrderBy(orderBy string) string {
	if orderBy != "" {
		if string(orderBy[0]) == "-" {
			orderBy = orderBy[1:] + " DESC"
		}
		orderBy = "ORDER BY " + orderBy
	}
	return orderBy
}

func (svc *ManagementService) updateBTCInDB(unixTime int64, lastValue string) {
	if err := svc.db.UpdateLastRecordForBTC(); err != nil {
		log.Printf("error with update last record for btc, err %s\n", err)
		return
	}
	tm := time.Unix(0, unixTime*int64(time.Millisecond))
	value, err := strconv.ParseFloat(lastValue, 64)
	if err != nil {
		log.Printf("error with parse float, err %s\n", err)
		return
	}
	btc := &models.BTC{
		ID:        uuid.NewV4().String(),
		Value:     value,
		CreatedAt: &tm,
		Latest:    true,
	}
	btcToFiat, err := svc.GetBTCToFiat(btc)
	if err != nil {
		log.Println(err)
		return
	}
	btc.CurrenciesToBTC, err = json.Marshal(btcToFiat)
	if err != nil {
		log.Println(err)
		return
	}
	if err = svc.db.CreateBTCRecord(btc); err != nil {
		log.Println(err)
		return
	}
	log.Println("BTC/USDT and BTC/Fiat updated in db")
}

func (svc *ManagementService) updateFiatInDB(data []byte) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = charset.NewReaderLabel
	var val ValCurs
	if err := decoder.Decode(&val); err != nil {
		log.Printf("error with decoding body, err: %s\n", err.Error())
		return
	}
	currencies, usdrub, err := serializeFiatCurrenciesData(val.Valutes)
	model := &models.Fiat{
		ID:     uuid.NewV4().String(),
		Latest: true,
		USDRUB: usdrub,
	}
	json.Unmarshal(currencies, &model.Currencies)
	date, err := svc.db.GetLastDateForFiat()
	if err != nil {
		log.Printf("error with decoding body, err: %s\n", err.Error())
		return
	}
	if date != nil {
		// if we already have data today -> skip
		d := date.Format(time.RFC3339[:10])
		t := time.Now().Format(time.RFC3339[:10])
		if d == t {
			log.Println("already have data for fiat today")
			return
		}
	}
	// if there is no data today -> create a new one
	// set old data as latest=false
	if err := svc.db.SetAllRecordsFiatLatestFalse(); err != nil {
		log.Printf("error with set latest false for all fiat records, err: %s\n", err.Error())
		return
	}
	// create a new record for fiat currencies
	if err = svc.db.CreateFiatRecord(model); err != nil {
		log.Printf("error with creating fiat record, err: %s\n", err.Error())
		return
	}
	log.Println("Fiat updated in db")
}

func (svc *ManagementService) GetLastBTC() (*models.BTC, error) {
	model, err := svc.db.GetLastBTC()
	if err != nil {
		return nil, fmt.Errorf("error to get last btcusdt, err: %s\n", err.Error())
	}
	return model, nil
}

func (svc *ManagementService) GetAllBTC(limit, offset int, orderBy string) ([]models.BTC, error) {
	orderBy = serializeOrderBy(orderBy)
	modelsData, err := svc.db.GetAllBTC(limit, offset, orderBy)
	if err != nil {
		return nil, fmt.Errorf("error to get all btcusdt data, err: %s\n", err.Error())
	}
	return modelsData, nil
}

func (svc *ManagementService) GetLastFiat() (*models.Fiat, error) {
	model, err := svc.db.GetLastFiat()
	if err != nil {
		return nil, fmt.Errorf("error to get last fiat, err: %s\n", err.Error())
	}
	return model, nil
}

func (svc *ManagementService) GetFiatHistory(limit, offset int, orderBy string) ([]models.Fiat, error) {
	orderBy = serializeOrderBy(orderBy)
	modelsData, err := svc.db.GetAllFiat(limit, offset, orderBy)
	if err != nil {
		return nil, fmt.Errorf("error to get all fiat history, err: %s\n", err.Error())
	}
	return modelsData, nil
}

func (svc *ManagementService) GetBTCToFiat(btc *models.BTC) (*map[string]float64, error) {
	lastFiat, err := svc.db.GetLastFiat()
	if err != nil {
		return nil, fmt.Errorf("error to get last fiat, err: %s\n", err.Error())
	}
	btcToFiat := make(map[string]float64, 34)
	var currencies []models.Currency
	json.Unmarshal(lastFiat.Currencies, &currencies)
	for _, c := range currencies {
		btcToFiat[c.CharCode] = btc.Value * lastFiat.USDRUB / c.Val * float64(c.Nominal)
	}
	return &btcToFiat, nil
}
