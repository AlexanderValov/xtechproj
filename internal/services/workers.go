package services

import (
	"encoding/json"
	"encoding/xml"
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
		log.Println(err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
	}
	go response.Body.Close()
	var r *BTCUSDTResponse
	if err := json.Unmarshal(body, &r); err != nil {
		log.Println(err)
	}
	if lastPrice != r.Data.Last {
		// create new record
		go svc.updateBTCInDB(r.Data.Time, r.Data.Last)
		lastPrice = r.Data.Last
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
	response, err := getResponse(svc.cfg.URLs.Fiat)
	if err != nil {
		log.Printf("error with getting response from %s, err: %s\n", svc.cfg.URLs.Fiat, err.Error())
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("error with reading response, err: %s\n", err.Error())
	}
	go svc.updateFiatInDB(data)
	go response.Body.Close()
}
