package main

import (
	"log"
	"os"

	"github.com/ryanaidilp/sulteng-procurement/backend/internal/config"
	"github.com/ryanaidilp/sulteng-procurement/backend/internal/scraper"
	"github.com/ryanaidilp/sulteng-procurement/backend/internal/storage"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	log.Printf("Starting scraper for year %d", cfg.ScraperYear)

	client := scraper.NewSirupClient(cfg.SirupURL, cfg.ScraperYear)
	records, err := client.FetchAll()
	if err != nil {
		log.Fatalf("fetch failed: %v", err)
	}
	log.Printf("Fetched %d records from SIRUP", len(records))

	store := storage.NewParquetStore(cfg.DataDir)
	domainRecords := scraper.ToDomain(records)

	if err := store.Write(cfg.ScraperYear, domainRecords); err != nil {
		log.Fatalf("write parquet failed: %v", err)
	}

	log.Printf("Successfully wrote %d records to parquet", len(domainRecords))
}
