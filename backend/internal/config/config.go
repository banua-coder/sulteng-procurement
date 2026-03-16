package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port         string
	DataDir      string
	SirupURL     string
	ScraperYear  int
	CronSchedule string
}

func Load() *Config {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	sirupURL := os.Getenv("SIRUP_URL")
	if sirupURL == "" {
		sirupURL = "https://sirup.inaproc.id/sirup/caripaketctr/search"
	}

	scraperYear := time.Now().Year()
	if v := os.Getenv("SCRAPER_YEAR"); v != "" {
		if y, err := strconv.Atoi(v); err == nil {
			scraperYear = y
		}
	}

	cronSchedule := os.Getenv("CRON_SCHEDULE")
	if cronSchedule == "" {
		cronSchedule = "0 2 * * *" // 2am daily
	}

	return &Config{
		Port:         port,
		DataDir:      dataDir,
		SirupURL:     sirupURL,
		ScraperYear:  scraperYear,
		CronSchedule: cronSchedule,
	}
}
