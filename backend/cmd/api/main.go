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
	currentData := data
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
		scraped := scraper.ToDomain(records)
		if err := store.Write(cfg.ScraperYear, scraped); err != nil {
			log.Printf("Cron: write failed: %v", err)
			return
		}
		mu.Lock()
		currentData = scraped
		mu.Unlock()
		handler.SetService(service.NewProcurementService(scraped))
		log.Printf("Cron: scrape complete, %d records loaded", len(scraped))
	}

	runSpseScrape := func() {
		log.Println("Cron: starting SPSE scrape")
		spseClient := scraper.NewSpseClient(cfg.SpseURL, cfg.ScraperYear)
		tenders, err := spseClient.FetchAll()
		if err != nil {
			log.Printf("Cron: SPSE FetchAll failed: %v", err)
			return
		}
		tenders = spseClient.EnrichWinners(tenders)
		if err := spseStore.Write(cfg.ScraperYear, tenders); err != nil {
			log.Printf("Cron: SPSE write failed: %v", err)
			return
		}
		mu.RLock()
		rup := currentData
		mu.RUnlock()
		handler.SetRealisasiService(service.NewRealisasiService(rup, tenders))
		log.Printf("Cron: SPSE scrape complete, %d tenders loaded", len(tenders))
	}

	c := cron.New()
	if _, err := c.AddFunc(cfg.CronSchedule, runScrape); err != nil {
		log.Fatalf("Invalid cron schedule %q: %v", cfg.CronSchedule, err)
	}
	if _, err := c.AddFunc(cfg.CronSchedule, runSpseScrape); err != nil {
		log.Printf("Warning: could not add SPSE cron: %v", err)
	}
	c.Start()
	defer c.Stop()

	log.Printf("API server listening on :%s (cron: %s)", cfg.Port, cfg.CronSchedule)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
