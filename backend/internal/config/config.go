package config

import "os"

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
	cronSchedule := os.Getenv("CRON_SCHEDULE")
	return &Config{
		Port:         port,
		DataDir:      dataDir,
		SirupURL:     "https://sirup.inaproc.id/sirup/caripaketctr/search",
		ScraperYear:  2026,
		CronSchedule: cronSchedule,
	}
}
