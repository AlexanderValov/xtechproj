package server

import (
	"encoding/json"
	"github.com/gorilla/schema"
	"log"
	"net/http"
	"time"
)

type lastBTCResponse struct {
	Value    float64    `json:"value"`
	Datetime *time.Time `json:"datetime"`
}

func (s *Server) LatestBTCUSDT(w http.ResponseWriter, r *http.Request) {
	model, err := s.service.GetLastBTC()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := lastBTCResponse{
		Value:    model.InUSDT,
		Datetime: model.CreatedAt,
	}
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type BTCHistoryResponse struct {
	Total   int          `json:"total"`
	History []BTCHistory `json:"history"`
}

type BTCHistory struct {
	Value  float64 `json:"value"`
	Date   string  `json:"date"`
	Latest bool    `json:"latest"`
}

func (s *Server) BTCUSDTWithHistory(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filter := new(Filter)
	if err := schema.NewDecoder().Decode(filter, r.Form); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	models, err := s.service.GetAllBTC(filter.Limit, filter.Offset, filter.OrderBy)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var history []BTCHistory
	for _, m := range models {
		history = append(history, BTCHistory{
			Value:  m.InUSDT,
			Date:   m.CreatedAt.Format(time.RFC3339[:19]),
			Latest: m.Latest,
		})
	}
	response := BTCHistoryResponse{
		Total:   len(history),
		History: history,
	}
	json.Marshal(response)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
