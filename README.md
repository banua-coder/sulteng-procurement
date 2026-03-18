# Sulteng Procurement Dashboard

Dashboard transparansi pengadaan barang dan jasa pemerintah Provinsi Sulawesi Tengah.

Data bersumber dari SIRUP LKPP (sirup.inaproc.id) mencakup 14 KLPD di Sulawesi Tengah.

## Stack

- **Backend:** Go 1.22+, Chi router, Parquet storage
- **Frontend:** Vue 3 + Vite + TypeScript + TailwindCSS + Chart.js
- **Infrastructure:** Docker Compose + Traefik v3

## Quick Start

```bash
# Run scraper to fetch data
docker compose run --rm scraper

# Start all services
docker compose up -d
```

Visit `http://localhost` to view the dashboard.
