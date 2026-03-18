#!/bin/bash
# Called by CI on the remote server after git pull
set -euo pipefail

echo "==> Building Docker images..."
docker compose build --no-cache

echo "==> Running scraper to refresh data..."
docker compose run --rm scraper || echo "Scraper failed, continuing with existing data"

echo "==> Starting services..."
docker compose up -d backend frontend traefik

echo "==> Cleaning up old images..."
docker image prune -f

echo "==> Deploy complete!"
docker compose ps
