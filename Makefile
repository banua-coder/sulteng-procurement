.PHONY: help dev backend frontend scrape test build

DATA_DIR ?= ./backend/data
API_PORT ?= 8080
SCRAPER_YEAR ?= 2026

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Run backend + frontend concurrently (requires terminal multiplexer or two shells)
	@echo "Starting backend on :$(API_PORT) and frontend on :5173"
	@$(MAKE) -j2 backend frontend

backend: ## Run Go API server (auto-reloads if air is installed, else plain go run)
	@mkdir -p $(DATA_DIR)
	@if command -v air >/dev/null 2>&1; then \
		cd backend && DATA_DIR=../$(DATA_DIR) API_PORT=$(API_PORT) SCRAPER_YEAR=$(SCRAPER_YEAR) air -c ../.air.toml; \
	else \
		cd backend && DATA_DIR=../$(DATA_DIR) API_PORT=$(API_PORT) SCRAPER_YEAR=$(SCRAPER_YEAR) go run ./cmd/api; \
	fi

frontend: ## Run Vite dev server on :5173
	cd frontend && npm run dev

scrape: ## Fetch latest data from SIRUP and write to parquet
	@mkdir -p $(DATA_DIR)
	cd backend && DATA_DIR=../$(DATA_DIR) SCRAPER_YEAR=$(SCRAPER_YEAR) go run ./cmd/scraper

scrape-spse: ## Fetch SPSE tender winner data and write to parquet
	@mkdir -p $(DATA_DIR)
	cd backend && DATA_DIR=../$(DATA_DIR) SCRAPER_YEAR=$(SCRAPER_YEAR) go run ./cmd/spse-scraper

test: ## Run all Go tests
	cd backend && go test ./... -v

build: ## Build Go binaries
	cd backend && go build ./cmd/api && go build ./cmd/scraper
	cd frontend && npm run build
