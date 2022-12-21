package main

import (
	"XTechProject/cmd/config"
	"XTechProject/internal/repository"
	"XTechProject/internal/server"
	"XTechProject/internal/services"
	"XTechProject/pkg/postgres"
	"log"
)

func main() {
	// init config
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("error with creating config, err: %s", err.Error())
	}
	// init postgres
	db, err := postgres.NewPostgresDB(cfg.DB.URL)
	if err != nil {
		log.Fatalf("error with starting postgres, err: %s", err.Error())
	}
	// init repository and create/check tables
	repo := repository.New(db)
	// init services and start workers
	service := services.NewManagementService(repo, cfg)
	// run workers
	go service.RunWorkers()
	//init server
	srv := server.NewServer(cfg.PORT, service)
	if err != nil {
		log.Fatalf("error with starting server, err: %s", err.Error())
	}
	// run server
	log.Println("Listening and serving: http://localhost:" + cfg.PORT)
	log.Panic(srv.ListenAndServe())
}
