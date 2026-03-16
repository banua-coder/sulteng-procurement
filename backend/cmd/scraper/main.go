package main

import (
	"fmt"

	"github.com/ryanaidilp/sulteng-procurement/backend/internal/config"
)

func main() {
	cfg := config.Load()
	fmt.Printf("Scraper starting for year %d\n", cfg.ScraperYear)
}
