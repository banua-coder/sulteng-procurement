package main

import (
	"log"
	"net/http"

	"github.com/banua-coder/sulteng-procurement/backend/internal/api"
	"github.com/banua-coder/sulteng-procurement/backend/internal/config"
	"github.com/banua-coder/sulteng-procurement/backend/internal/service"
	"github.com/banua-coder/sulteng-procurement/backend/internal/storage"
)

func main() {
	cfg := config.Load()

	store := storage.NewParquetStore(cfg.DataDir)
	data, err := store.Read(cfg.ScraperYear)
	if err != nil {
		log.Printf("Warning: could not load data: %v (run scraper first)", err)
		data = nil
	}
	log.Printf("Loaded %d procurement records", len(data))

	svc := service.NewProcurementService(data)
	handler := api.NewHandler(svc)
	router := api.NewRouter(handler)

	log.Printf("API server listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
