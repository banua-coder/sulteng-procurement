package main

import (
	"log"
	"os"

	"github.com/banua-coder/sulteng-procurement/backend/internal/config"
	"github.com/banua-coder/sulteng-procurement/backend/internal/scraper"
	"github.com/banua-coder/sulteng-procurement/backend/internal/storage"
)

func main() {
	cfg := config.Load()

	spseBase := os.Getenv("SPSE_URL")
	if spseBase == "" {
		spseBase = "https://spse.inaproc.id/sultengprov"
	}

	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Fatalf("mkdir: %v", err)
	}

	client := scraper.NewSpseClient(spseBase, cfg.ScraperYear)

	log.Printf("Fetching SPSE data for year %d from %s", cfg.ScraperYear, spseBase)
	records, err := client.FetchAll()
	if err != nil {
		log.Fatalf("FetchAll: %v", err)
	}
	log.Printf("Fetched %d tender records, enriching winners...", len(records))

	records = client.EnrichWinners(records)

	store := storage.NewSpseStore(cfg.DataDir)
	if err := store.Write(cfg.ScraperYear, records); err != nil {
		log.Fatalf("Write: %v", err)
	}
	log.Printf("Wrote %d SPSE records to parquet", len(records))
}
