package server

import (
	"encoding/json"
	"log"
	"net/http"
)

func (s *Server) LastBTCFiat(w http.ResponseWriter, r *http.Request) {
	btc, err := s.service.GetLastBTC()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(&btc.BTCToFiat); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
