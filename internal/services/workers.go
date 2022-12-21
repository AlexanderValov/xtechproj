package services

import (
	"XTechProject/internal/models"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"golang.org/x/net/html/charset"
	"io"
	"io/ioutil"
	"log"
)

type BTCUSDTResponse struct {
	Code string `json:"code"`
	Data struct {
		Time             int64  `json:"time"`
		Symbol           string `json:"symbol"`
		Buy              string `json:"buy"`
		Sell             string `json:"sell"`
		ChangeRate       string `json:"changeRate"`
		ChangePrice      string `json:"changePrice"`
		High             string `json:"high"`
		Low              string `json:"low"`
		Vol              string `json:"vol"`
		VolValue         string `json:"volValue"`
		Last             string `json:"last"`
		AveragePrice     string `json:"averagePrice"`
		TakerFeeRate     string `json:"takerFeeRate"`
		MakerFeeRate     string `json:"makerFeeRate"`
		TakerCoefficient string `json:"takerCoefficient"`
		MakerCoefficient string `json:"makerCoefficient"`
	} `json:"data"`
}

var lastPrice string

func (svc *ManagementService) BTCWorker() {
	log.Println("BTCWorker triggered")
	response, err := getResponse(svc.cfg.URLs.BTCUSDT)
	if err != nil {
		log.Printf("BTCWorker: error in getResponse, err: %s", err.Error())
		return
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("BTCWorker: error in io.ReadAll, err: %s", err.Error())
		return
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Printf("error in response.Body.Close(), err: %s", err.Error())
		}
	}()
	var r *BTCUSDTResponse
	if err := json.Unmarshal(body, &r); err != nil {
		log.Printf("BTCWorker: error in json.Unmarshal, err: %s", err.Error())
	}
	if lastPrice != r.Data.Last {
		lastPrice = r.Data.Last
		// create new record
		go svc.UpdateBTCInDB(r.Data.Time, r.Data.Last)
	}
}

type (
	ValCurs struct {
		XMLName xml.Name `xml:"ValCurs"`
		Date    string   `xml:"Date,attr"`
		Name    string   `xml:"name,attr"`
		Valutes []Valute `xml:"Valute"`
	}
	Valute struct {
		ID       string `xml:"ID,attr"`
		NumCode  string `xml:"NumCode"`
		CharCode string `xml:"CharCode"`
		Nominal  string `xml:"Nominal"`
		Name     string `xml:"Name"`
		Value    string `xml:"Value"`
	}
)

func (svc *ManagementService) FiatWorker() {
	log.Println("FiatWorker triggered")
	// if there is data today -> stop
	if err := svc.CheckLastDateUpdatingFiatCurrencies(); err != nil {
		log.Printf("FiatWorker: error in checkLastDateUpdatingFiatCurrencies: %s", err.Error())
		return
	}
	response, err := getResponse(svc.cfg.URLs.Fiat)
	if err != nil {
		log.Printf("FiatWorker: error in getResponse from %s, err: %s\n", svc.cfg.URLs.Fiat, err.Error())
		return
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Printf("error in response.Body.Close(), err: %s", err.Error())
		}
	}()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("FiatWorker: error in ioutil.ReadAll, err: %s\n", err.Error())
		return
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = charset.NewReaderLabel
	var val ValCurs
	if err := decoder.Decode(&val); err != nil {
		log.Printf("FiatWorker: error in decoder.Decode, err: %s\n", err.Error())
		return
	}
	currencies, usdrub, err := serializeFiatCurrenciesData(val.Valutes)
	model := &models.Fiat{
		Latest: true,
		USDRUB: usdrub,
	}
	if err = json.Unmarshal(currencies, &model.Currencies); err != nil {
		log.Printf("FiatWorker: error in json.Unmarshal, err: %s\n", err.Error())
		return
	}
	// set old data as latest=false
	if err := svc.db.SetAllRecordsFiatLatestFalse(); err != nil {
		log.Printf("FiatWorker: error in SetAllRecordsFiatLatestFalse, err: %s\n", err.Error())
		return
	}
	// create a new record for fiat currencies
	if err = svc.db.CreateFiatRecord(model); err != nil {
		log.Printf("FiatWorker:error in CreateFiatRecord, err: %s\n", err.Error())
		return
	}
	log.Println("Fiat updated in db")
}
