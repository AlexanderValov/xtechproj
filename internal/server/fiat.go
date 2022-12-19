package server

import (
	"XTechProject/internal/models"
	"encoding/json"
	"github.com/gorilla/schema"
	"log"
	"net/http"
	"time"
)

type lastFiatResponse struct {
	Date    string          `json:"date"`
	Valutes json.RawMessage `json:"valutes"`
}

func (s *Server) LastFiat(w http.ResponseWriter, r *http.Request) {
	model, err := s.service.GetLastFiat()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := lastFiatResponse{
		Date:    model.CreatedAt.Format(time.RFC3339[:10]),
		Valutes: model.Currencies,
	}
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type FiatHistoryResponse struct {
	Total   int                      `json:"total"`
	History []map[string]interface{} `json:"history"`
}

func (s *Server) FiatHistory(w http.ResponseWriter, r *http.Request) {
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
	modelsData, err := s.service.GetFiatHistory(filter.Limit, filter.Offset, filter.OrderBy)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	history := make([]map[string]interface{}, 0, len(modelsData)*34)
	for _, m := range modelsData {
		var currencies []models.Currency
		json.Unmarshal(m.Currencies, &currencies)
		body := make(map[string]interface{}, 34)
		for _, v := range currencies {
			body[v.CharCode] = v.Val
		}
		body["date"] = m.CreatedAt.Format(time.RFC3339[:10])
		body["latest"] = m.Latest
		history = append(history, body)
	}
	response := FiatHistoryResponse{
		Total:   len(history),
		History: history,
	}
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
