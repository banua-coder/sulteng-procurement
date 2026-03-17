package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/robfig/cron/v3"

	"github.com/banua-coder/sulteng-procurement/backend/internal/api"
	"github.com/banua-coder/sulteng-procurement/backend/internal/config"
	"github.com/banua-coder/sulteng-procurement/backend/internal/scraper"
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

	var mu sync.RWMutex
	svc := service.NewProcurementService(data)
	handler := api.NewHandler(svc)

	spseStore := storage.NewSpseStore(cfg.DataDir)
	if spseStore.Exists(cfg.ScraperYear) {
		tenders, err := spseStore.Read(cfg.ScraperYear)
		if err != nil {
			log.Printf("Warning: could not load SPSE data: %v", err)
		} else {
			log.Printf("Loaded %d SPSE tender records", len(tenders))
			realSvc := service.NewRealisasiService(data, tenders)
			handler.SetRealisasiService(realSvc)
		}
	}

	router := api.NewRouter(handler)

	client := scraper.NewSirupClient(cfg.SirupURL, cfg.ScraperYear)

	runScrape := func() {
		log.Println("Cron: starting daily scrape")
		records, err := client.FetchAll()
		if err != nil {
			log.Printf("Cron: scrape failed: %v", err)
			return
		}
		domain := scraper.ToDomain(records)
		if err := store.Write(cfg.ScraperYear, domain); err != nil {
			log.Printf("Cron: write failed: %v", err)
			return
		}
		mu.Lock()
		newSvc := service.NewProcurementService(domain)
		handler.SetService(newSvc)
		mu.Unlock()
		log.Printf("Cron: scrape complete, %d records loaded", len(domain))
	}

	c := cron.New()
	if _, err := c.AddFunc(cfg.CronSchedule, runScrape); err != nil {
		log.Fatalf("Invalid cron schedule %q: %v", cfg.CronSchedule, err)
	}
	c.Start()
	defer c.Stop()

	log.Printf("API server listening on :%s (cron: %s)", cfg.Port, cfg.CronSchedule)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
