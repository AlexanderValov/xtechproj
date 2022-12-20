package server

import (
	"XTechProject/internal/services"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type (
	Server struct {
		*http.Server
		service services.Servicer
	}
	Filter struct {
		Offset  int    `schema:"offset"`
		Limit   int    `schema:"limit"`
		OrderBy string `schema:"order_by"`
	}
)

func NewServer(port string, service *services.ManagementService) *Server {
	srv := &Server{
		service: service,
	}

	srv.Server = &http.Server{
		Addr:           ":" + port,
		Handler:        srv.Handler(),
		MaxHeaderBytes: 1 << 20, // 1 MB
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}
	return srv
}

func (s *Server) Handler() *mux.Router {
	r := mux.NewRouter()

	router := r.PathPrefix("/api").Subrouter()

	router.HandleFunc("/btcusdt", s.LatestBTCUSDT).Methods(http.MethodGet)
	router.HandleFunc("/btcusdt", s.BTCUSDTWithHistory).Methods(http.MethodPost)

	router.HandleFunc("/currencies", s.LastFiat).Methods(http.MethodGet)
	router.HandleFunc("/currencies", s.FiatHistory).Methods(http.MethodPost)

	router.HandleFunc("/latest", s.LastBTCFiat).Methods(http.MethodGet)

	return r
}
