# Sulawesi Tengah Procurement Dashboard — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a procurement transparency dashboard for Sulawesi Tengah province government, scraping data from SIRUP/INAPROC and presenting it via a Go backend + Vue frontend monorepo deployed with Docker/Traefik.

**Architecture:** Go scraper fetches data from SIRUP API (`sirup.inaproc.id/sirup/caripaketctr/search`) for all 14 Sulawesi Tengah KLPD codes (~39K packages), stores as Parquet files for efficient columnar queries. Go API serves aggregated + paginated data. Vue 3 frontend renders a dashboard with summary cards, bar chart, top-5 list, and sortable/filterable/paginated table. Docker Compose orchestrates all services behind Traefik reverse proxy.

**Tech Stack:**
- **Backend:** Go 1.22+, Chi router, parquet-go (xitongsys/parquet-go or apache/arrow), cron scheduler
- **Frontend:** Vue 3 + Vite, TypeScript, Composition API `<script setup>`, TailwindCSS, Chart.js (via vue-chartjs), Heroicons
- **Infrastructure:** Docker, Docker Compose, Traefik v3, multi-stage builds
- **CI/CD:** GitHub Actions with [banua-coder-workflow](https://github.com/banua-coder/banua-coder-workflow) reusable workflows
- **Data:** Apache Parquet for storage, daily cron scraper

---

## Data Source Reference

**SIRUP Search API:**
```
GET https://sirup.inaproc.id/sirup/caripaketctr/search
```

**Query Parameters (DataTables server-side):**
| Param | Description | Example |
|-------|-------------|---------|
| `tahunAnggaran` | Budget year | `2026` |
| `kldi` | KLPD code(s), comma-separated | `569,575,570` |
| `jenisPengadaan` | Procurement type filter | (empty = all) |
| `metodePengadaan` | Method filter | (empty = all) |
| `minPagu` / `maxPagu` | Budget range | |
| `bulan` | Month filter | |
| `lokasi` | Location filter | |
| `pdn` / `ukm` | Domestic product / small business | |
| `draw` | DataTables draw counter | `1` |
| `start` | Offset | `0` |
| `length` | Page size (max appears to be 100) | `100` |
| `search[value]` | Text search | |
| `order[0][column]` / `order[0][dir]` | Sort | `5`, `DESC` |

**Response (JSON):**
```json
{
  "recordsTotal": 39065,
  "recordsFiltered": 39065,
  "data": [
    {
      "id": 62953279,
      "paket": "Package name",
      "pagu": 10700040,
      "jenisPengadaan": "Jasa Lainnya",
      "metode": "Pengadaan Langsung",
      "pemilihan": "January 2026",
      "satuanKerja": "Sekretariat Daerah",
      "kldi": "Kab. Banggai",
      "lokasi": "Sulawesi Tengah, Banggai (Kab.)",
      "sumberDana": "APBD",
      "isPDN": true,
      "isUMK": true,
      "idBulan": 1,
      "idKldi": "D427",
      "id_referensi": 570,
      "idSatker": 182394,
      "idMetode": 8,
      "idJenisPengadaan": 4,
      "idsLokasi": "14926",
      "pds": false
    }
  ]
}
```

**Sulawesi Tengah KLPD Codes (14 regions):**

| Code | Name |
|------|------|
| 569 | Provinsi Sulawesi Tengah |
| 575 | Kota Palu |
| 570 | Kab. Banggai |
| 571 | Kab. Banggai Kepulauan |
| 578 | Kab. Banggai Laut |
| 572 | Kab. Buol |
| 187 | Kab. Donggala |
| 573 | Kab. Morowali |
| 579 | Kab. Morowali Utara |
| 576 | Kab. Parigi Moutong |
| 574 | Kab. Poso |
| 577 | Kab. Sigi |
| 782 | Kab. Tojo Una Una |
| 787 | Kab. Toli Toli |

**Detail endpoint (HTML modal):**
```
GET https://sirup.inaproc.id/sirup/rup/detailPaketPenyedia2020?idPaket={id}
```
Returns HTML with fields: Provinsi, Kabupaten/Kota, Detail Lokasi, Sumber Dana, T.A., KLPD, MAK, Pagu, Jenis Pengadaan, timeline dates.

---

## Monorepo Structure

```
sulteng-procurement/
├── docs/plans/
├── backend/
│   ├── cmd/
│   │   ├── api/main.go            # API server entrypoint
│   │   └── scraper/main.go        # Scraper CLI entrypoint
│   ├── internal/
│   │   ├── config/config.go       # Env-based configuration
│   │   ├── domain/
│   │   │   └── procurement.go     # Domain types
│   │   ├── scraper/
│   │   │   ├── sirup.go           # SIRUP API client
│   │   │   └── sirup_test.go
│   │   ├── storage/
│   │   │   ├── parquet.go         # Parquet read/write
│   │   │   └── parquet_test.go
│   │   ├── api/
│   │   │   ├── handler.go         # HTTP handlers
│   │   │   ├── handler_test.go
│   │   │   └── router.go          # Chi router setup
│   │   └── service/
│   │       ├── procurement.go     # Business logic / aggregation
│   │       └── procurement_test.go
│   ├── data/                      # Parquet files (gitignored)
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── App.vue
│   │   ├── main.ts
│   │   ├── api/
│   │   │   └── procurement.ts     # API client
│   │   ├── composables/
│   │   │   └── useProcurement.ts  # Data fetching composable
│   │   ├── components/
│   │   │   ├── SummaryCards.vue
│   │   │   ├── CategoryChart.vue
│   │   │   ├── TopProcurements.vue
│   │   │   ├── DataTable.vue
│   │   │   └── FilterBar.vue
│   │   └── types/
│   │       └── procurement.ts     # TypeScript interfaces
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   └── Dockerfile
├── docker-compose.yml
├── .env.example
└── .gitignore
```

---

## Phase 1: Project Scaffolding & Domain Types

### Task 1: Initialize monorepo and Git

**Files:**
- Create: `sulteng-procurement/.gitignore`
- Create: `sulteng-procurement/README.md` (minimal)

**Step 1: Initialize git repo**
```bash
cd /Users/ryanaidilp/Documents/Projects/Web/sulteng-procurement
git init
git checkout -b develop
```

**Step 2: Create .gitignore**
```gitignore
# Go
backend/data/*.parquet
backend/tmp/

# Node
frontend/node_modules/
frontend/dist/

# IDE
.idea/
.vscode/
*.swp

# Env
.env
.env.local

# OS
.DS_Store
```

**Step 3: Commit**
```bash
git add .gitignore docs/
git commit -m "chore: init monorepo with project plan"
```

---

### Task 2: Initialize Go backend module

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/api/main.go` (stub)
- Create: `backend/cmd/scraper/main.go` (stub)
- Create: `backend/internal/config/config.go`

**Step 1: Initialize Go module**
```bash
cd backend
go mod init github.com/ryanaidilp/sulteng-procurement/backend
```

**Step 2: Create config**
```go
// backend/internal/config/config.go
package config

import "os"

type Config struct {
	Port        string
	DataDir     string
	SirupURL    string
	ScraperYear int
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
	return &Config{
		Port:        port,
		DataDir:     dataDir,
		SirupURL:    "https://sirup.inaproc.id/sirup/caripaketctr/search",
		ScraperYear: 2026,
	}
}
```

**Step 3: Create API main stub**
```go
// backend/cmd/api/main.go
package main

import (
	"fmt"
	"github.com/ryanaidilp/sulteng-procurement/backend/internal/config"
)

func main() {
	cfg := config.Load()
	fmt.Printf("API server starting on port %s\n", cfg.Port)
}
```

**Step 4: Create scraper main stub**
```go
// backend/cmd/scraper/main.go
package main

import (
	"fmt"
	"github.com/ryanaidilp/sulteng-procurement/backend/internal/config"
)

func main() {
	cfg := config.Load()
	fmt.Printf("Scraper starting for year %d\n", cfg.ScraperYear)
}
```

**Step 5: Commit**
```bash
git add backend/
git commit -m "chore: scaffold Go backend module with config"
```

---

### Task 3: Define domain types

**Files:**
- Create: `backend/internal/domain/procurement.go`

**Step 1: Write domain types**

These map directly to the SIRUP API response fields plus computed aggregation types.

```go
// backend/internal/domain/procurement.go
package domain

type Procurement struct {
	ID              int64   `json:"id" parquet:"name=id, type=INT64"`
	Paket           string  `json:"paket" parquet:"name=paket, type=BYTE_ARRAY, convertedtype=UTF8"`
	Pagu            float64 `json:"pagu" parquet:"name=pagu, type=DOUBLE"`
	JenisPengadaan  string  `json:"jenisPengadaan" parquet:"name=jenis_pengadaan, type=BYTE_ARRAY, convertedtype=UTF8"`
	Metode          string  `json:"metode" parquet:"name=metode, type=BYTE_ARRAY, convertedtype=UTF8"`
	Pemilihan       string  `json:"pemilihan" parquet:"name=pemilihan, type=BYTE_ARRAY, convertedtype=UTF8"`
	SatuanKerja     string  `json:"satuanKerja" parquet:"name=satuan_kerja, type=BYTE_ARRAY, convertedtype=UTF8"`
	KLDI            string  `json:"kldi" parquet:"name=kldi, type=BYTE_ARRAY, convertedtype=UTF8"`
	Lokasi          string  `json:"lokasi" parquet:"name=lokasi, type=BYTE_ARRAY, convertedtype=UTF8"`
	SumberDana      string  `json:"sumberDana" parquet:"name=sumber_dana, type=BYTE_ARRAY, convertedtype=UTF8"`
	IsPDN           bool    `json:"isPDN" parquet:"name=is_pdn, type=BOOLEAN"`
	IsUMK           bool    `json:"isUMK" parquet:"name=is_umk, type=BOOLEAN"`
	IdBulan         int     `json:"idBulan" parquet:"name=id_bulan, type=INT32"`
	IdKldi          string  `json:"idKldi" parquet:"name=id_kldi, type=BYTE_ARRAY, convertedtype=UTF8"`
	IdReferensi     int     `json:"idReferensi" parquet:"name=id_referensi, type=INT32"`
	IdSatker        int     `json:"idSatker" parquet:"name=id_satker, type=INT32"`
	IdMetode        int     `json:"idMetode" parquet:"name=id_metode, type=INT32"`
	IdJenisPengadaan int    `json:"idJenisPengadaan" parquet:"name=id_jenis_pengadaan, type=INT32"`
}

type Summary struct {
	TotalPagu       float64          `json:"totalPagu"`
	TotalPaket      int              `json:"totalPaket"`
	JenisCount      int              `json:"jenisCount"`
	KLDICount       int              `json:"kldiCount"`
	TopKLDI         string           `json:"topKldi"`
	ByJenis         []CategoryTotal  `json:"byJenis"`
	ByKLDI          []CategoryTotal  `json:"byKldi"`
	ByMetode        []CategoryTotal  `json:"byMetode"`
}

type CategoryTotal struct {
	Name  string  `json:"name"`
	Total float64 `json:"total"`
	Count int     `json:"count"`
}

type ProcurementQuery struct {
	Page            int    `json:"page"`
	PageSize        int    `json:"pageSize"`
	Search          string `json:"search"`
	KLDI            string `json:"kldi"`
	JenisPengadaan  string `json:"jenisPengadaan"`
	Metode          string `json:"metode"`
	SortBy          string `json:"sortBy"`
	SortDir         string `json:"sortDir"`
}

type PaginatedResult struct {
	Data         []Procurement `json:"data"`
	Total        int           `json:"total"`
	Page         int           `json:"page"`
	PageSize     int           `json:"pageSize"`
	TotalPages   int           `json:"totalPages"`
}
```

**Step 2: Commit**
```bash
git add backend/internal/domain/
git commit -m "feat: define procurement domain types with parquet tags"
```

---

## Phase 2: Scraper

### Task 4: Implement SIRUP API client

**Files:**
- Create: `backend/internal/scraper/sirup.go`
- Create: `backend/internal/scraper/sirup_test.go`

**Step 1: Write the test**
```go
// backend/internal/scraper/sirup_test.go
package scraper

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("tahunAnggaran") != "2026" {
			t.Errorf("expected tahunAnggaran=2026, got %s", q.Get("tahunAnggaran"))
		}
		if q.Get("kldi") == "" {
			t.Error("expected kldi to be set")
		}
		resp := SirupResponse{
			RecordsTotal:    2,
			RecordsFiltered: 2,
			Data: []SirupRecord{
				{ID: 1, Paket: "Test Package 1", Pagu: 1000000, JenisPengadaan: "Barang", Metode: "Tender", KLDI: "Kab. Poso"},
				{ID: 2, Paket: "Test Package 2", Pagu: 2000000, JenisPengadaan: "Jasa Lainnya", Metode: "E-Purchasing", KLDI: "Kota Palu"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewSirupClient(server.URL, 2026)
	records, total, err := client.FetchPage(0, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total=2, got %d", total)
	}
	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}
	if records[0].Paket != "Test Package 1" {
		t.Errorf("expected 'Test Package 1', got %s", records[0].Paket)
	}
}

func TestFetchAllPages(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := SirupResponse{RecordsTotal: 150, RecordsFiltered: 150}
		q := r.URL.Query()
		start := q.Get("start")
		if start == "0" {
			for i := 0; i < 100; i++ {
				resp.Data = append(resp.Data, SirupRecord{ID: int64(i), Paket: "Pkg", Pagu: 1000})
			}
		} else {
			for i := 100; i < 150; i++ {
				resp.Data = append(resp.Data, SirupRecord{ID: int64(i), Paket: "Pkg", Pagu: 1000})
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewSirupClient(server.URL, 2026)
	records, err := client.FetchAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 150 {
		t.Errorf("expected 150 records, got %d", len(records))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}
```

**Step 2: Run test to verify it fails**
```bash
cd backend && go test ./internal/scraper/ -v
```
Expected: FAIL — types and functions not defined.

**Step 3: Implement SIRUP client**
```go
// backend/internal/scraper/sirup.go
package scraper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ryanaidilp/sulteng-procurement/backend/internal/domain"
)

var SultengKLDICodes = []int{569, 575, 570, 571, 578, 572, 187, 573, 579, 576, 574, 577, 782, 787}

type SirupRecord struct {
	ID              int64   `json:"id"`
	Paket           string  `json:"paket"`
	Pagu            float64 `json:"pagu"`
	JenisPengadaan  string  `json:"jenisPengadaan"`
	Metode          string  `json:"metode"`
	Pemilihan       string  `json:"pemilihan"`
	SatuanKerja     string  `json:"satuanKerja"`
	KLDI            string  `json:"kldi"`
	Lokasi          string  `json:"lokasi"`
	SumberDana      string  `json:"sumberDana"`
	IsPDN           bool    `json:"isPDN"`
	IsUMK           bool    `json:"isUMK"`
	IdBulan         int     `json:"idBulan"`
	IdKldi          string  `json:"idKldi"`
	IdReferensi     int     `json:"id_referensi"`
	IdSatker        int     `json:"idSatker"`
	IdMetode        int     `json:"idMetode"`
	IdJenisPengadaan int    `json:"idJenisPengadaan"`
}

type SirupResponse struct {
	RecordsTotal    int           `json:"recordsTotal"`
	RecordsFiltered int           `json:"recordsFiltered"`
	Data            []SirupRecord `json:"data"`
}

type SirupClient struct {
	baseURL string
	year    int
	http    *http.Client
}

func NewSirupClient(baseURL string, year int) *SirupClient {
	return &SirupClient{
		baseURL: baseURL,
		year:    year,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *SirupClient) kldiParam() string {
	parts := make([]string, len(SultengKLDICodes))
	for i, code := range SultengKLDICodes {
		parts[i] = fmt.Sprintf("%d", code)
	}
	return strings.Join(parts, ",")
}

func (c *SirupClient) FetchPage(start, length int) ([]SirupRecord, int, error) {
	url := fmt.Sprintf(
		"%s?tahunAnggaran=%d&jenisPengadaan=&metodePengadaan=&minPagu=&maxPagu=&bulan=&lokasi=&kldi=%s&pdn=&ukm=&draw=1&start=%d&length=%d&search[value]=&order[0][column]=5&order[0][dir]=DESC",
		c.baseURL, c.year, c.kldiParam(), start, length,
	)

	resp, err := c.http.Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch page start=%d: %w", start, err)
	}
	defer resp.Body.Close()

	var result SirupResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("decode response start=%d: %w", start, err)
	}

	return result.Data, result.RecordsTotal, nil
}

func (c *SirupClient) FetchAll() ([]SirupRecord, error) {
	pageSize := 100
	var all []SirupRecord

	first, total, err := c.FetchPage(0, pageSize)
	if err != nil {
		return nil, err
	}
	all = append(all, first...)
	log.Printf("Total records: %d, fetched first page: %d", total, len(first))

	for start := pageSize; start < total; start += pageSize {
		records, _, err := c.FetchPage(start, pageSize)
		if err != nil {
			return nil, err
		}
		all = append(all, records...)
		log.Printf("Fetched %d/%d records", len(all), total)
		time.Sleep(200 * time.Millisecond) // rate limiting
	}

	return all, nil
}

func ToDomain(records []SirupRecord) []domain.Procurement {
	result := make([]domain.Procurement, len(records))
	for i, r := range records {
		result[i] = domain.Procurement{
			ID:               r.ID,
			Paket:            r.Paket,
			Pagu:             r.Pagu,
			JenisPengadaan:   r.JenisPengadaan,
			Metode:           r.Metode,
			Pemilihan:        r.Pemilihan,
			SatuanKerja:      r.SatuanKerja,
			KLDI:             r.KLDI,
			Lokasi:           r.Lokasi,
			SumberDana:       r.SumberDana,
			IsPDN:            r.IsPDN,
			IsUMK:            r.IsUMK,
			IdBulan:          r.IdBulan,
			IdKldi:           r.IdKldi,
			IdReferensi:      r.IdReferensi,
			IdSatker:         r.IdSatker,
			IdMetode:         r.IdMetode,
			IdJenisPengadaan: r.IdJenisPengadaan,
		}
	}
	return result
}
```

**Step 4: Run tests**
```bash
cd backend && go test ./internal/scraper/ -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add backend/internal/scraper/
git commit -m "feat: implement SIRUP API client with pagination"
```

---

### Task 5: Implement Parquet storage

**Files:**
- Create: `backend/internal/storage/parquet.go`
- Create: `backend/internal/storage/parquet_test.go`

**Step 1: Add parquet dependency**
```bash
cd backend && go get github.com/xitongsys/parquet-go@latest github.com/xitongsys/parquet-go-source@latest
```

**Step 2: Write the test**
```go
// backend/internal/storage/parquet_test.go
package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ryanaidilp/sulteng-procurement/backend/internal/domain"
)

func TestWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	store := NewParquetStore(dir)

	records := []domain.Procurement{
		{ID: 1, Paket: "Package A", Pagu: 1000000, JenisPengadaan: "Barang", KLDI: "Kota Palu", Metode: "Tender"},
		{ID: 2, Paket: "Package B", Pagu: 2000000, JenisPengadaan: "Jasa Lainnya", KLDI: "Kab. Poso", Metode: "E-Purchasing"},
	}

	err := store.Write(2026, records)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	path := filepath.Join(dir, "procurement_2026.parquet")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("parquet file was not created")
	}

	loaded, err := store.Read(2026)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if len(loaded) != 2 {
		t.Errorf("expected 2 records, got %d", len(loaded))
	}
	if loaded[0].Paket != "Package A" {
		t.Errorf("expected 'Package A', got %s", loaded[0].Paket)
	}
	if loaded[1].Pagu != 2000000 {
		t.Errorf("expected pagu=2000000, got %f", loaded[1].Pagu)
	}
}
```

**Step 3: Run test to verify it fails**
```bash
cd backend && go test ./internal/storage/ -v
```
Expected: FAIL

**Step 4: Implement parquet storage**
```go
// backend/internal/storage/parquet.go
package storage

import (
	"fmt"
	"path/filepath"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/writer"

	"github.com/ryanaidilp/sulteng-procurement/backend/internal/domain"
)

type ParquetStore struct {
	dir string
}

func NewParquetStore(dir string) *ParquetStore {
	return &ParquetStore{dir: dir}
}

func (s *ParquetStore) path(year int) string {
	return filepath.Join(s.dir, fmt.Sprintf("procurement_%d.parquet", year))
}

func (s *ParquetStore) Write(year int, records []domain.Procurement) error {
	fw, err := local.NewLocalFileWriter(s.path(year))
	if err != nil {
		return fmt.Errorf("create file writer: %w", err)
	}

	pw, err := writer.NewParquetWriter(fw, new(domain.Procurement), 4)
	if err != nil {
		return fmt.Errorf("create parquet writer: %w", err)
	}
	pw.RowGroupSize = 128 * 1024 * 1024
	pw.CompressionType = 1 // SNAPPY

	for _, r := range records {
		if err := pw.Write(r); err != nil {
			return fmt.Errorf("write record: %w", err)
		}
	}

	if err := pw.WriteStop(); err != nil {
		return fmt.Errorf("finalize parquet: %w", err)
	}
	return fw.Close()
}

func (s *ParquetStore) Read(year int) ([]domain.Procurement, error) {
	fr, err := local.NewLocalFileReader(s.path(year))
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer fr.Close()

	pr, err := reader.NewParquetReader(fr, new(domain.Procurement), 4)
	if err != nil {
		return nil, fmt.Errorf("create reader: %w", err)
	}
	defer pr.ReadStop()

	num := int(pr.GetNumRows())
	records := make([]domain.Procurement, num)
	if err := pr.Read(&records); err != nil {
		return nil, fmt.Errorf("read records: %w", err)
	}

	return records, nil
}

func (s *ParquetStore) Exists(year int) bool {
	_, err := local.NewLocalFileReader(s.path(year))
	return err == nil
}
```

**Step 5: Run tests**
```bash
cd backend && go test ./internal/storage/ -v
```
Expected: PASS

**Step 6: Commit**
```bash
git add backend/
git commit -m "feat: implement parquet storage for procurement data"
```

---

### Task 6: Wire up scraper CLI

**Files:**
- Modify: `backend/cmd/scraper/main.go`

**Step 1: Implement full scraper entrypoint**
```go
// backend/cmd/scraper/main.go
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
```

**Step 2: Build and test manually**
```bash
cd backend && go build ./cmd/scraper/
```
Expected: builds successfully.

**Step 3: Commit**
```bash
git add backend/cmd/scraper/
git commit -m "feat: wire up scraper CLI to fetch and store procurement data"
```

---

## Phase 3: API Server

### Task 7: Implement procurement service (aggregation + query)

**Files:**
- Create: `backend/internal/service/procurement.go`
- Create: `backend/internal/service/procurement_test.go`

**Step 1: Write tests**
```go
// backend/internal/service/procurement_test.go
package service

import (
	"testing"

	"github.com/ryanaidilp/sulteng-procurement/backend/internal/domain"
)

func sampleData() []domain.Procurement {
	return []domain.Procurement{
		{ID: 1, Paket: "Pembangunan Jalan", Pagu: 5000000000, JenisPengadaan: "Pekerjaan Konstruksi", KLDI: "Kota Palu", Metode: "Tender", SatuanKerja: "Dinas PU"},
		{ID: 2, Paket: "Pengadaan Komputer", Pagu: 100000000, JenisPengadaan: "Barang", KLDI: "Kab. Poso", Metode: "E-Purchasing", SatuanKerja: "Dinas Kominfo"},
		{ID: 3, Paket: "Jasa Konsultan", Pagu: 250000000, JenisPengadaan: "Jasa Konsultansi", KLDI: "Kota Palu", Metode: "Seleksi", SatuanKerja: "Bappeda"},
		{ID: 4, Paket: "Catering Rapat", Pagu: 50000000, JenisPengadaan: "Jasa Lainnya", KLDI: "Kab. Banggai", Metode: "Pengadaan Langsung", SatuanKerja: "Setda"},
		{ID: 5, Paket: "Renovasi Kantor", Pagu: 3000000000, JenisPengadaan: "Pekerjaan Konstruksi", KLDI: "Kota Palu", Metode: "Tender", SatuanKerja: "Dinas PU"},
	}
}

func TestGetSummary(t *testing.T) {
	svc := NewProcurementService(sampleData())
	summary := svc.GetSummary()

	if summary.TotalPaket != 5 {
		t.Errorf("expected 5 paket, got %d", summary.TotalPaket)
	}
	expectedTotal := 8400000000.0
	if summary.TotalPagu != expectedTotal {
		t.Errorf("expected totalPagu=%f, got %f", expectedTotal, summary.TotalPagu)
	}
	if summary.JenisCount != 4 {
		t.Errorf("expected 4 jenis, got %d", summary.JenisCount)
	}
	if summary.KLDICount != 3 {
		t.Errorf("expected 3 kldi, got %d", summary.KLDICount)
	}
	if summary.TopKLDI != "Kota Palu" {
		t.Errorf("expected topKLDI='Kota Palu', got %s", summary.TopKLDI)
	}
}

func TestQueryPagination(t *testing.T) {
	svc := NewProcurementService(sampleData())
	result := svc.Query(domain.ProcurementQuery{Page: 1, PageSize: 2, SortBy: "pagu", SortDir: "desc"})

	if result.Total != 5 {
		t.Errorf("expected total=5, got %d", result.Total)
	}
	if len(result.Data) != 2 {
		t.Errorf("expected 2 items, got %d", len(result.Data))
	}
	if result.Data[0].Pagu < result.Data[1].Pagu {
		t.Error("expected descending sort by pagu")
	}
	if result.TotalPages != 3 {
		t.Errorf("expected 3 pages, got %d", result.TotalPages)
	}
}

func TestQuerySearch(t *testing.T) {
	svc := NewProcurementService(sampleData())
	result := svc.Query(domain.ProcurementQuery{Page: 1, PageSize: 10, Search: "jalan"})

	if result.Total != 1 {
		t.Errorf("expected 1 result for 'jalan', got %d", result.Total)
	}
}

func TestQueryFilterByKLDI(t *testing.T) {
	svc := NewProcurementService(sampleData())
	result := svc.Query(domain.ProcurementQuery{Page: 1, PageSize: 10, KLDI: "Kota Palu"})

	if result.Total != 3 {
		t.Errorf("expected 3 results for Kota Palu, got %d", result.Total)
	}
}
```

**Step 2: Run tests — verify fail**
```bash
cd backend && go test ./internal/service/ -v
```

**Step 3: Implement service**
```go
// backend/internal/service/procurement.go
package service

import (
	"math"
	"sort"
	"strings"

	"github.com/ryanaidilp/sulteng-procurement/backend/internal/domain"
)

type ProcurementService struct {
	data []domain.Procurement
}

func NewProcurementService(data []domain.Procurement) *ProcurementService {
	return &ProcurementService{data: data}
}

func (s *ProcurementService) SetData(data []domain.Procurement) {
	s.data = data
}

func (s *ProcurementService) GetSummary() domain.Summary {
	var totalPagu float64
	jenis := map[string]float64{}
	jenisCount := map[string]int{}
	kldi := map[string]float64{}
	kldiCount := map[string]int{}
	metode := map[string]float64{}
	metodeCount := map[string]int{}

	for _, p := range s.data {
		totalPagu += p.Pagu
		jenis[p.JenisPengadaan] += p.Pagu
		jenisCount[p.JenisPengadaan]++
		kldi[p.KLDI] += p.Pagu
		kldiCount[p.KLDI]++
		metode[p.Metode] += p.Pagu
		metodeCount[p.Metode]++
	}

	byJenis := toSortedCategoryTotals(jenis, jenisCount)
	byKLDI := toSortedCategoryTotals(kldi, kldiCount)
	byMetode := toSortedCategoryTotals(metode, metodeCount)

	topKLDI := ""
	if len(byKLDI) > 0 {
		topKLDI = byKLDI[0].Name
	}

	return domain.Summary{
		TotalPagu:  totalPagu,
		TotalPaket: len(s.data),
		JenisCount: len(jenis),
		KLDICount:  len(kldi),
		TopKLDI:    topKLDI,
		ByJenis:    byJenis,
		ByKLDI:     byKLDI,
		ByMetode:   byMetode,
	}
}

func (s *ProcurementService) Query(q domain.ProcurementQuery) domain.PaginatedResult {
	filtered := s.filter(q)
	s.sortData(filtered, q.SortBy, q.SortDir)

	total := len(filtered)
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 25
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	return domain.PaginatedResult{
		Data:       filtered[start:end],
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

func (s *ProcurementService) GetFilters() map[string][]string {
	kldiSet := map[string]bool{}
	jenisSet := map[string]bool{}
	metodeSet := map[string]bool{}

	for _, p := range s.data {
		kldiSet[p.KLDI] = true
		jenisSet[p.JenisPengadaan] = true
		metodeSet[p.Metode] = true
	}

	return map[string][]string{
		"kldi":           toSortedKeys(kldiSet),
		"jenisPengadaan": toSortedKeys(jenisSet),
		"metode":         toSortedKeys(metodeSet),
	}
}

func (s *ProcurementService) filter(q domain.ProcurementQuery) []domain.Procurement {
	var result []domain.Procurement
	search := strings.ToLower(q.Search)

	for _, p := range s.data {
		if q.KLDI != "" && p.KLDI != q.KLDI {
			continue
		}
		if q.JenisPengadaan != "" && p.JenisPengadaan != q.JenisPengadaan {
			continue
		}
		if q.Metode != "" && p.Metode != q.Metode {
			continue
		}
		if search != "" {
			haystack := strings.ToLower(p.Paket + " " + p.SatuanKerja + " " + p.KLDI + " " + p.Lokasi)
			if !strings.Contains(haystack, search) {
				continue
			}
		}
		result = append(result, p)
	}
	return result
}

func (s *ProcurementService) sortData(data []domain.Procurement, sortBy, sortDir string) {
	desc := strings.ToLower(sortDir) == "desc"
	sort.Slice(data, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "paket":
			less = data[i].Paket < data[j].Paket
		case "kldi":
			less = data[i].KLDI < data[j].KLDI
		case "jenisPengadaan":
			less = data[i].JenisPengadaan < data[j].JenisPengadaan
		default: // pagu
			less = data[i].Pagu < data[j].Pagu
		}
		if desc {
			return !less
		}
		return less
	})
}

func toSortedCategoryTotals(totals map[string]float64, counts map[string]int) []domain.CategoryTotal {
	result := make([]domain.CategoryTotal, 0, len(totals))
	for name, total := range totals {
		result = append(result, domain.CategoryTotal{Name: name, Total: total, Count: counts[name]})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Total > result[j].Total
	})
	return result
}

func toSortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
```

**Step 4: Run tests**
```bash
cd backend && go test ./internal/service/ -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add backend/internal/service/
git commit -m "feat: implement procurement service with aggregation and query"
```

---

### Task 8: Implement HTTP API handlers + router

**Files:**
- Create: `backend/internal/api/router.go`
- Create: `backend/internal/api/handler.go`
- Create: `backend/internal/api/handler_test.go`

**Step 1: Write handler tests**
```go
// backend/internal/api/handler_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ryanaidilp/sulteng-procurement/backend/internal/domain"
	"github.com/ryanaidilp/sulteng-procurement/backend/internal/service"
)

func setupHandler() *Handler {
	data := []domain.Procurement{
		{ID: 1, Paket: "Build Road", Pagu: 5e9, JenisPengadaan: "Konstruksi", KLDI: "Kota Palu", Metode: "Tender"},
		{ID: 2, Paket: "Buy Laptop", Pagu: 1e8, JenisPengadaan: "Barang", KLDI: "Kab. Poso", Metode: "E-Purchasing"},
	}
	svc := service.NewProcurementService(data)
	return NewHandler(svc)
}

func TestGetSummary(t *testing.T) {
	h := setupHandler()
	r := NewRouter(h)

	req := httptest.NewRequest("GET", "/api/v1/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var summary domain.Summary
	json.NewDecoder(w.Body).Decode(&summary)
	if summary.TotalPaket != 2 {
		t.Errorf("expected 2 paket, got %d", summary.TotalPaket)
	}
}

func TestGetProcurements(t *testing.T) {
	h := setupHandler()
	r := NewRouter(h)

	req := httptest.NewRequest("GET", "/api/v1/procurements?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result domain.PaginatedResult
	json.NewDecoder(w.Body).Decode(&result)
	if result.Total != 2 {
		t.Errorf("expected total=2, got %d", result.Total)
	}
}

func TestGetFilters(t *testing.T) {
	h := setupHandler()
	r := NewRouter(h)

	req := httptest.NewRequest("GET", "/api/v1/filters", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var filters map[string][]string
	json.NewDecoder(w.Body).Decode(&filters)
	if len(filters["kldi"]) != 2 {
		t.Errorf("expected 2 kldi options, got %d", len(filters["kldi"]))
	}
}
```

**Step 2: Run tests — verify fail**
```bash
cd backend && go test ./internal/api/ -v
```

**Step 3: Implement handler and router**
```go
// backend/internal/api/handler.go
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ryanaidilp/sulteng-procurement/backend/internal/domain"
	"github.com/ryanaidilp/sulteng-procurement/backend/internal/service"
)

type Handler struct {
	svc *service.ProcurementService
}

func NewHandler(svc *service.ProcurementService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.svc.GetSummary())
}

func (h *Handler) GetProcurements(w http.ResponseWriter, r *http.Request) {
	q := domain.ProcurementQuery{
		Page:           intParam(r, "page", 1),
		PageSize:       intParam(r, "pageSize", 25),
		Search:         r.URL.Query().Get("search"),
		KLDI:           r.URL.Query().Get("kldi"),
		JenisPengadaan: r.URL.Query().Get("jenisPengadaan"),
		Metode:         r.URL.Query().Get("metode"),
		SortBy:         r.URL.Query().Get("sortBy"),
		SortDir:        r.URL.Query().Get("sortDir"),
	}
	writeJSON(w, h.svc.Query(q))
}

func (h *Handler) GetFilters(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.svc.GetFilters())
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func intParam(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
```

```go
// backend/internal/api/router.go
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/summary", h.GetSummary)
		r.Get("/procurements", h.GetProcurements)
		r.Get("/filters", h.GetFilters)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	return r
}
```

**Step 4: Add chi dependency and run tests**
```bash
cd backend && go get github.com/go-chi/chi/v5 github.com/go-chi/cors
go test ./internal/api/ -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add backend/
git commit -m "feat: implement HTTP API with summary, procurements, and filters endpoints"
```

---

### Task 9: Wire up API server main

**Files:**
- Modify: `backend/cmd/api/main.go`

**Step 1: Implement API entrypoint**
```go
// backend/cmd/api/main.go
package main

import (
	"log"
	"net/http"

	"github.com/ryanaidilp/sulteng-procurement/backend/internal/api"
	"github.com/ryanaidilp/sulteng-procurement/backend/internal/config"
	"github.com/ryanaidilp/sulteng-procurement/backend/internal/service"
	"github.com/ryanaidilp/sulteng-procurement/backend/internal/storage"
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

	svc := service.NewProcurementService(data)
	handler := api.NewHandler(svc)
	router := api.NewRouter(handler)

	log.Printf("API server listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
```

**Step 2: Build**
```bash
cd backend && go build ./cmd/api/
```
Expected: builds successfully.

**Step 3: Commit**
```bash
git add backend/cmd/api/
git commit -m "feat: wire up API server with parquet data loading"
```

---

## Phase 4: Vue Frontend

### Task 10: Scaffold Vue 3 + Vite + TypeScript + TailwindCSS

**Step 1: Create Vue project**
```bash
cd /Users/ryanaidilp/Documents/Projects/Web/sulteng-procurement
npm create vite@latest frontend -- --template vue-ts
cd frontend
npm install
npm install -D tailwindcss @tailwindcss/vite
npm install chart.js vue-chartjs
npm install @heroicons/vue
```

**Step 2: Configure Tailwind in vite.config.ts**
```ts
// frontend/vite.config.ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
```

**Step 3: Add Tailwind to main CSS**
```css
/* frontend/src/style.css */
@import "tailwindcss";
```

**Step 4: Commit**
```bash
git add frontend/
git commit -m "chore: scaffold Vue 3 + Vite + TypeScript + TailwindCSS frontend"
```

---

### Task 11: Create TypeScript types and API client

**Files:**
- Create: `frontend/src/types/procurement.ts`
- Create: `frontend/src/api/procurement.ts`

**Step 1: Define types**
```ts
// frontend/src/types/procurement.ts
export interface Procurement {
  id: number
  paket: string
  pagu: number
  jenisPengadaan: string
  metode: string
  pemilihan: string
  satuanKerja: string
  kldi: string
  lokasi: string
  sumberDana: string
  isPDN: boolean
  isUMK: boolean
}

export interface CategoryTotal {
  name: string
  total: number
  count: number
}

export interface Summary {
  totalPagu: number
  totalPaket: number
  jenisCount: number
  kldiCount: number
  topKldi: string
  byJenis: CategoryTotal[]
  byKldi: CategoryTotal[]
  byMetode: CategoryTotal[]
}

export interface PaginatedResult {
  data: Procurement[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export interface Filters {
  kldi: string[]
  jenisPengadaan: string[]
  metode: string[]
}

export interface QueryParams {
  page: number
  pageSize: number
  search: string
  kldi: string
  jenisPengadaan: string
  metode: string
  sortBy: string
  sortDir: string
}
```

**Step 2: Implement API client**
```ts
// frontend/src/api/procurement.ts
import type { Summary, PaginatedResult, Filters, QueryParams } from '../types/procurement'

const BASE = '/api/v1'

async function fetchJSON<T>(url: string): Promise<T> {
  const res = await fetch(url)
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  return res.json()
}

export function getSummary(): Promise<Summary> {
  return fetchJSON(`${BASE}/summary`)
}

export function getProcurements(params: Partial<QueryParams>): Promise<PaginatedResult> {
  const qs = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') {
      qs.set(key, String(value))
    }
  }
  return fetchJSON(`${BASE}/procurements?${qs}`)
}

export function getFilters(): Promise<Filters> {
  return fetchJSON(`${BASE}/filters`)
}
```

**Step 3: Commit**
```bash
git add frontend/src/types/ frontend/src/api/
git commit -m "feat: add TypeScript types and API client for procurement data"
```

---

### Task 12: Create useProcurement composable

**Files:**
- Create: `frontend/src/composables/useProcurement.ts`

**Step 1: Implement composable**
```ts
// frontend/src/composables/useProcurement.ts
import { ref, reactive, watch } from 'vue'
import { getSummary, getProcurements, getFilters } from '../api/procurement'
import type { Summary, PaginatedResult, Filters, QueryParams } from '../types/procurement'

export function useProcurement() {
  const summary = ref<Summary | null>(null)
  const result = ref<PaginatedResult | null>(null)
  const filters = ref<Filters | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const query = reactive<QueryParams>({
    page: 1,
    pageSize: 25,
    search: '',
    kldi: '',
    jenisPengadaan: '',
    metode: '',
    sortBy: 'pagu',
    sortDir: 'desc',
  })

  async function loadSummary() {
    try {
      summary.value = await getSummary()
    } catch (e: any) {
      error.value = e.message
    }
  }

  async function loadFilters() {
    try {
      filters.value = await getFilters()
    } catch (e: any) {
      error.value = e.message
    }
  }

  async function loadData() {
    loading.value = true
    error.value = null
    try {
      result.value = await getProcurements({ ...query })
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  function setPage(page: number) {
    query.page = page
  }

  function setSort(sortBy: string) {
    if (query.sortBy === sortBy) {
      query.sortDir = query.sortDir === 'desc' ? 'asc' : 'desc'
    } else {
      query.sortBy = sortBy
      query.sortDir = 'desc'
    }
  }

  function resetFilters() {
    query.page = 1
    query.search = ''
    query.kldi = ''
    query.jenisPengadaan = ''
    query.metode = ''
  }

  // Auto-reload when query changes
  watch(query, () => {
    loadData()
  })

  return {
    summary,
    result,
    filters,
    loading,
    error,
    query,
    loadSummary,
    loadFilters,
    loadData,
    setPage,
    setSort,
    resetFilters,
  }
}
```

**Step 2: Commit**
```bash
git add frontend/src/composables/
git commit -m "feat: add useProcurement composable for data management"
```

---

### Task 13: Build FilterBar component

**Files:**
- Create: `frontend/src/components/FilterBar.vue`

**Step 1: Implement component**
```vue
<!-- frontend/src/components/FilterBar.vue -->
<script setup lang="ts">
import type { Filters } from '../types/procurement'

const props = defineProps<{
  filters: Filters | null
  kldi: string
  jenisPengadaan: string
  metode: string
  search: string
  pageSize: number
}>()

const emit = defineEmits<{
  'update:kldi': [value: string]
  'update:jenisPengadaan': [value: string]
  'update:metode': [value: string]
  'update:search': [value: string]
  'update:pageSize': [value: number]
}>()
</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
    <div>
      <label class="block text-sm font-medium text-stone-600 mb-1">Wilayah (KLPD)</label>
      <select
        :value="kldi"
        class="w-full rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm"
        @change="emit('update:kldi', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">Semua wilayah</option>
        <option v-for="k in filters?.kldi" :key="k" :value="k">{{ k }}</option>
      </select>
    </div>
    <div>
      <label class="block text-sm font-medium text-stone-600 mb-1">Jenis pengadaan</label>
      <select
        :value="jenisPengadaan"
        class="w-full rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm"
        @change="emit('update:jenisPengadaan', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">Semua jenis</option>
        <option v-for="j in filters?.jenisPengadaan" :key="j" :value="j">{{ j }}</option>
      </select>
    </div>
    <div>
      <label class="block text-sm font-medium text-stone-600 mb-1">Metode pengadaan</label>
      <select
        :value="metode"
        class="w-full rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm"
        @change="emit('update:metode', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">Semua metode</option>
        <option v-for="m in filters?.metode" :key="m" :value="m">{{ m }}</option>
      </select>
    </div>
    <div>
      <label class="block text-sm font-medium text-stone-600 mb-1">Cari paket / satker</label>
      <input
        :value="search"
        type="text"
        placeholder="Ketik kata kunci..."
        class="w-full rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm"
        @input="emit('update:search', ($event.target as HTMLInputElement).value)"
      />
    </div>
    <div>
      <label class="block text-sm font-medium text-stone-600 mb-1">Baris per halaman</label>
      <select
        :value="pageSize"
        class="w-full rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm"
        @change="emit('update:pageSize', Number(($event.target as HTMLSelectElement).value))"
      >
        <option :value="10">10</option>
        <option :value="25">25</option>
        <option :value="50">50</option>
        <option :value="100">100</option>
      </select>
    </div>
  </div>
</template>
```

**Step 2: Commit**
```bash
git add frontend/src/components/FilterBar.vue
git commit -m "feat: add FilterBar component with dropdowns and search"
```

---

### Task 14: Build SummaryCards component

**Files:**
- Create: `frontend/src/components/SummaryCards.vue`

**Step 1: Implement component**
```vue
<!-- frontend/src/components/SummaryCards.vue -->
<script setup lang="ts">
import type { Summary } from '../types/procurement'

defineProps<{ summary: Summary | null }>()

function formatRupiah(value: number): string {
  if (value >= 1e12) return `Rp${(value / 1e12).toFixed(2)} T`
  if (value >= 1e9) return `Rp${(value / 1e9).toFixed(2)} M`
  if (value >= 1e6) return `Rp${(value / 1e6).toFixed(1)} Jt`
  return `Rp${value.toLocaleString('id-ID')}`
}
</script>

<template>
  <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
    <div class="rounded-xl bg-amber-50 border border-amber-200 p-4">
      <div class="text-sm text-stone-500">Total pagu anggaran</div>
      <div class="text-2xl font-bold text-stone-800 mt-1">
        {{ summary ? formatRupiah(summary.totalPagu) : '-' }}
      </div>
    </div>
    <div class="rounded-xl bg-amber-50 border border-amber-200 p-4">
      <div class="text-sm text-stone-500">Jumlah paket</div>
      <div class="text-2xl font-bold text-stone-800 mt-1">
        {{ summary ? summary.totalPaket.toLocaleString('id-ID') : '-' }}
      </div>
    </div>
    <div class="rounded-xl bg-amber-50 border border-amber-200 p-4">
      <div class="text-sm text-stone-500">Jumlah wilayah</div>
      <div class="text-2xl font-bold text-stone-800 mt-1">
        {{ summary?.kldiCount ?? '-' }}
      </div>
    </div>
    <div class="rounded-xl bg-amber-50 border border-amber-200 p-4">
      <div class="text-sm text-stone-500">Wilayah terbesar</div>
      <div class="text-2xl font-bold text-stone-800 mt-1 truncate">
        {{ summary?.topKldi ?? '-' }}
      </div>
    </div>
  </div>
</template>
```

**Step 2: Commit**
```bash
git add frontend/src/components/SummaryCards.vue
git commit -m "feat: add SummaryCards component with formatted rupiah values"
```

---

### Task 15: Build CategoryChart component

**Files:**
- Create: `frontend/src/components/CategoryChart.vue`

**Step 1: Implement component**
```vue
<!-- frontend/src/components/CategoryChart.vue -->
<script setup lang="ts">
import { computed } from 'vue'
import { Bar } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
} from 'chart.js'
import type { CategoryTotal } from '../types/procurement'

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip)

const props = defineProps<{
  data: CategoryTotal[]
  title: string
  maxItems?: number
}>()

const chartData = computed(() => {
  const items = props.data.slice(0, props.maxItems ?? 10)
  return {
    labels: items.map((d) => d.name),
    datasets: [
      {
        data: items.map((d) => d.total),
        backgroundColor: [
          '#292524', '#44403c', '#57534e', '#78716c',
          '#a8a29e', '#d6d3d1', '#e7e5e4', '#f5f5f4',
          '#fafaf9', '#b8b2a8',
        ],
        borderRadius: 4,
      },
    ],
  }
})

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (ctx: any) => {
          const val = ctx.raw as number
          if (val >= 1e12) return `Rp${(val / 1e12).toFixed(2)} T`
          if (val >= 1e9) return `Rp${(val / 1e9).toFixed(2)} M`
          return `Rp${(val / 1e6).toFixed(1)} Jt`
        },
      },
    },
  },
  scales: {
    y: {
      ticks: {
        callback: (val: number) => {
          if (val >= 1e12) return `Rp${(val / 1e12).toFixed(1)}T`
          if (val >= 1e9) return `Rp${(val / 1e9).toFixed(0)}M`
          return `Rp${(val / 1e6).toFixed(0)}Jt`
        },
      },
    },
    x: {
      ticks: {
        maxRotation: 45,
        font: { size: 10 },
      },
    },
  },
}))
</script>

<template>
  <div>
    <h3 class="font-semibold text-stone-800">{{ title }}</h3>
    <p class="text-sm text-stone-500 mb-4">
      Grafik menampilkan {{ maxItems ?? 10 }} teratas berdasarkan total pagu.
    </p>
    <div class="h-72">
      <Bar :data="chartData" :options="chartOptions" />
    </div>
  </div>
</template>
```

**Step 2: Commit**
```bash
git add frontend/src/components/CategoryChart.vue
git commit -m "feat: add CategoryChart component with bar chart visualization"
```

---

### Task 16: Build TopProcurements component

**Files:**
- Create: `frontend/src/components/TopProcurements.vue`

**Step 1: Implement component**
```vue
<!-- frontend/src/components/TopProcurements.vue -->
<script setup lang="ts">
import type { Procurement } from '../types/procurement'

defineProps<{ items: Procurement[] }>()

function formatRupiah(value: number): string {
  if (value >= 1e12) return `Rp${(value / 1e12).toFixed(2)} T`
  if (value >= 1e9) return `Rp${(value / 1e9).toFixed(1)} M`
  if (value >= 1e6) return `Rp${(value / 1e6).toFixed(1)} Jt`
  return `Rp${value.toLocaleString('id-ID')}`
}
</script>

<template>
  <div>
    <h3 class="font-semibold text-stone-800">Paket pengadaan terbesar</h3>
    <p class="text-sm text-stone-500 mb-4">5 paket dengan pagu tertinggi.</p>
    <div class="space-y-4">
      <div v-for="item in items.slice(0, 5)" :key="item.id" class="border-b border-stone-200 pb-3">
        <div class="font-medium text-stone-800 text-sm leading-snug">{{ item.paket }}</div>
        <span class="inline-block mt-1 text-xs px-2 py-0.5 rounded-full bg-amber-100 text-amber-800 font-medium">
          {{ item.jenisPengadaan }}
        </span>
        <div class="text-xs text-stone-500 mt-1">{{ item.satuanKerja }}</div>
        <div class="text-sm font-semibold text-stone-700 mt-1">{{ formatRupiah(item.pagu) }}</div>
      </div>
    </div>
  </div>
</template>
```

**Step 2: Commit**
```bash
git add frontend/src/components/TopProcurements.vue
git commit -m "feat: add TopProcurements component showing top 5 by pagu"
```

---

### Task 17: Build DataTable component

**Files:**
- Create: `frontend/src/components/DataTable.vue`

**Step 1: Implement component**
```vue
<!-- frontend/src/components/DataTable.vue -->
<script setup lang="ts">
import type { PaginatedResult } from '../types/procurement'

defineProps<{
  result: PaginatedResult | null
  loading: boolean
  sortBy: string
  sortDir: string
}>()

const emit = defineEmits<{
  sort: [field: string]
  page: [page: number]
}>()

function formatRupiah(value: number): string {
  return `Rp${value.toLocaleString('id-ID')}`
}

function sortIcon(field: string, currentSort: string, dir: string): string {
  if (field !== currentSort) return ''
  return dir === 'desc' ? ' \u25BC' : ' \u25B2'
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-2">
      <h3 class="font-semibold text-stone-800">Tabel detail</h3>
      <div class="text-sm text-stone-500" v-if="result">
        Menampilkan {{ result.data.length }} dari {{ result.total.toLocaleString('id-ID') }} data
        <span class="ml-4">Halaman {{ result.page }} dari {{ result.totalPages }}</span>
      </div>
    </div>

    <div class="overflow-x-auto rounded-lg border border-stone-200">
      <table class="min-w-full text-sm">
        <thead class="bg-stone-50 text-stone-600">
          <tr>
            <th
              v-for="col in [
                { key: 'kldi', label: 'Wilayah' },
                { key: 'satuanKerja', label: 'Satuan Kerja' },
                { key: 'paket', label: 'Paket' },
                { key: 'jenisPengadaan', label: 'Jenis' },
                { key: 'metode', label: 'Metode' },
                { key: 'pagu', label: 'Pagu' },
              ]"
              :key="col.key"
              class="px-3 py-2 text-left cursor-pointer hover:bg-stone-100 whitespace-nowrap"
              @click="emit('sort', col.key)"
            >
              {{ col.label }}{{ sortIcon(col.key, sortBy, sortDir) }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="6" class="px-3 py-8 text-center text-stone-400">Memuat data...</td>
          </tr>
          <tr v-else-if="!result?.data.length">
            <td colspan="6" class="px-3 py-8 text-center text-stone-400">Tidak ada data</td>
          </tr>
          <tr
            v-for="item in result?.data"
            :key="item.id"
            class="border-t border-stone-100 hover:bg-stone-50"
          >
            <td class="px-3 py-2">{{ item.kldi }}</td>
            <td class="px-3 py-2">{{ item.satuanKerja }}</td>
            <td class="px-3 py-2 max-w-xs truncate">{{ item.paket }}</td>
            <td class="px-3 py-2 whitespace-nowrap">{{ item.jenisPengadaan }}</td>
            <td class="px-3 py-2 whitespace-nowrap">{{ item.metode }}</td>
            <td class="px-3 py-2 text-right whitespace-nowrap font-mono">{{ formatRupiah(item.pagu) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="flex justify-end gap-2 mt-3" v-if="result && result.totalPages > 1">
      <button
        class="px-3 py-1.5 text-sm rounded-lg border border-stone-300 hover:bg-stone-100 disabled:opacity-40"
        :disabled="result.page <= 1"
        @click="emit('page', result!.page - 1)"
      >
        Sebelumnya
      </button>
      <button
        class="px-3 py-1.5 text-sm rounded-lg border border-stone-300 hover:bg-stone-100 disabled:opacity-40"
        :disabled="result.page >= result.totalPages"
        @click="emit('page', result!.page + 1)"
      >
        Berikutnya
      </button>
    </div>
  </div>
</template>
```

**Step 2: Commit**
```bash
git add frontend/src/components/DataTable.vue
git commit -m "feat: add DataTable component with sorting and pagination"
```

---

### Task 18: Compose App.vue dashboard

**Files:**
- Modify: `frontend/src/App.vue`

**Step 1: Implement main dashboard**
```vue
<!-- frontend/src/App.vue -->
<script setup lang="ts">
import { onMounted, computed, watch } from 'vue'
import { useProcurement } from './composables/useProcurement'
import FilterBar from './components/FilterBar.vue'
import SummaryCards from './components/SummaryCards.vue'
import CategoryChart from './components/CategoryChart.vue'
import TopProcurements from './components/TopProcurements.vue'
import DataTable from './components/DataTable.vue'

const {
  summary,
  result,
  filters,
  loading,
  error,
  query,
  loadSummary,
  loadFilters,
  loadData,
  setPage,
  setSort,
} = useProcurement()

const topItems = computed(() => {
  if (!result.value?.data) return []
  return [...result.value.data].sort((a, b) => b.pagu - a.pagu).slice(0, 5)
})

// Debounce search
let searchTimeout: ReturnType<typeof setTimeout>
function onSearchUpdate(val: string) {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    query.search = val
    query.page = 1
  }, 300)
}

function onFilterChange(key: 'kldi' | 'jenisPengadaan' | 'metode', val: string) {
  query[key] = val
  query.page = 1
}

onMounted(() => {
  loadSummary()
  loadFilters()
  loadData()
})
</script>

<template>
  <div class="min-h-screen bg-stone-100">
    <div class="max-w-7xl mx-auto px-4 py-8">
      <h1 class="text-3xl font-bold text-stone-800 mb-1">
        Anggaran Pengadaan Sulawesi Tengah 2026
      </h1>
      <p class="text-stone-500 mb-6">
        Dashboard pengadaan barang dan jasa pemerintah provinsi dan kabupaten/kota di Sulawesi Tengah.
      </p>

      <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
        {{ error }}
      </div>

      <FilterBar
        :filters="filters"
        :kldi="query.kldi"
        :jenis-pengadaan="query.jenisPengadaan"
        :metode="query.metode"
        :search="query.search"
        :page-size="query.pageSize"
        @update:kldi="onFilterChange('kldi', $event)"
        @update:jenis-pengadaan="onFilterChange('jenisPengadaan', $event)"
        @update:metode="onFilterChange('metode', $event)"
        @update:search="onSearchUpdate"
        @update:page-size="query.pageSize = $event; query.page = 1"
      />

      <div class="mt-6">
        <SummaryCards :summary="summary" />
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-5 gap-6 mt-6">
        <div class="lg:col-span-3 bg-white rounded-xl border border-stone-200 p-5">
          <CategoryChart
            v-if="summary?.byKldi"
            :data="summary.byKldi"
            title="Pagu anggaran per wilayah"
            :max-items="10"
          />
        </div>
        <div class="lg:col-span-2 bg-white rounded-xl border border-stone-200 p-5">
          <TopProcurements :items="topItems" />
        </div>
      </div>

      <div class="mt-6 bg-white rounded-xl border border-stone-200 p-5">
        <DataTable
          :result="result"
          :loading="loading"
          :sort-by="query.sortBy"
          :sort-dir="query.sortDir"
          @sort="setSort"
          @page="setPage"
        />
      </div>

      <footer class="mt-8 text-xs text-stone-400">
        <p><strong>Sumber:</strong> SIRUP LKPP (sirup.inaproc.id) — Data RUP Sulawesi Tengah 2026</p>
      </footer>
    </div>
  </div>
</template>
```

**Step 2: Commit**
```bash
git add frontend/src/App.vue
git commit -m "feat: compose main dashboard with all components"
```

---

## Phase 5: Docker & Traefik Deployment

### Task 19: Backend Dockerfile

**Files:**
- Create: `backend/Dockerfile`

**Step 1: Create multi-stage Dockerfile**
```dockerfile
# backend/Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api
RUN CGO_ENABLED=0 go build -o /scraper ./cmd/scraper

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /api /app/api
COPY --from=builder /scraper /app/scraper
RUN mkdir -p /app/data
EXPOSE 8080
CMD ["/app/api"]
```

**Step 2: Commit**
```bash
git add backend/Dockerfile
git commit -m "chore: add multi-stage backend Dockerfile"
```

---

### Task 20: Frontend Dockerfile

**Files:**
- Create: `frontend/Dockerfile`
- Create: `frontend/nginx.conf`

**Step 1: Create nginx config**
```nginx
# frontend/nginx.conf
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

**Step 2: Create multi-stage Dockerfile**
```dockerfile
# frontend/Dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

**Step 3: Commit**
```bash
git add frontend/Dockerfile frontend/nginx.conf
git commit -m "chore: add frontend Dockerfile with nginx proxy"
```

---

### Task 21: Docker Compose with Traefik

**Files:**
- Create: `docker-compose.yml`
- Create: `.env.example`

**Step 1: Create env example**
```env
# .env.example
DOMAIN=procurement.sulteng.local
API_PORT=8080
DATA_DIR=/app/data
SCRAPER_YEAR=2026
```

**Step 2: Create Docker Compose**
```yaml
# docker-compose.yml
services:
  traefik:
    image: traefik:v3.0
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
    ports:
      - "80:80"
      - "8081:8080"  # Traefik dashboard
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro

  backend:
    build: ./backend
    environment:
      - API_PORT=8080
      - DATA_DIR=/app/data
      - SCRAPER_YEAR=2026
    volumes:
      - procurement-data:/app/data
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.api.rule=Host(`${DOMAIN:-localhost}`) && PathPrefix(`/api`)"
      - "traefik.http.routers.api.entrypoints=web"
      - "traefik.http.services.api.loadbalancer.server.port=8080"

  frontend:
    build: ./frontend
    depends_on:
      - backend
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.frontend.rule=Host(`${DOMAIN:-localhost}`)"
      - "traefik.http.routers.frontend.entrypoints=web"
      - "traefik.http.services.frontend.loadbalancer.server.port=80"

  scraper:
    build: ./backend
    command: ["/app/scraper"]
    environment:
      - DATA_DIR=/app/data
      - SCRAPER_YEAR=2026
      - SIRUP_URL=https://sirup.inaproc.id/sirup/caripaketctr/search
    volumes:
      - procurement-data:/app/data
    profiles:
      - scrape

volumes:
  procurement-data:
```

**Step 3: Commit**
```bash
git add docker-compose.yml .env.example
git commit -m "feat: add Docker Compose with Traefik reverse proxy"
```

---

## Phase 6: Cron Scraper + Polish

### Task 22: Add cron-based daily scraper option to backend

**Files:**
- Modify: `backend/cmd/api/main.go` — add optional cron schedule

**Step 1: Add robfig/cron dependency**
```bash
cd backend && go get github.com/robfig/cron/v3
```

**Step 2: Update API main to include optional cron**

Add to `backend/internal/config/config.go`:
```go
// Add field to Config
CronSchedule string  // e.g. "0 2 * * *" for 2am daily
```
```go
// In Load()
cronSchedule := os.Getenv("CRON_SCHEDULE")
```

Add cron setup in `backend/cmd/api/main.go` after loading data:
```go
if cfg.CronSchedule != "" {
    c := cron.New()
    c.AddFunc(cfg.CronSchedule, func() {
        log.Println("Cron: starting data refresh")
        client := scraper.NewSirupClient(cfg.SirupURL, cfg.ScraperYear)
        records, err := client.FetchAll()
        if err != nil {
            log.Printf("Cron scrape error: %v", err)
            return
        }
        domainRecords := scraper.ToDomain(records)
        if err := store.Write(cfg.ScraperYear, domainRecords); err != nil {
            log.Printf("Cron write error: %v", err)
            return
        }
        svc.SetData(domainRecords)
        log.Printf("Cron: refreshed %d records", len(domainRecords))
    })
    c.Start()
    log.Printf("Cron scheduled: %s", cfg.CronSchedule)
}
```

**Step 3: Commit**
```bash
git add backend/
git commit -m "feat: add optional cron-based daily data refresh"
```

---

### Task 23: Run all tests and verify

**Step 1: Run all Go tests**
```bash
cd backend && go test ./... -v
```
Expected: All PASS

**Step 2: Build all Docker images**
```bash
cd /Users/ryanaidilp/Documents/Projects/Web/sulteng-procurement
docker compose build
```
Expected: builds successfully.

**Step 3: Final commit with any fixes**
```bash
git add -A
git commit -m "chore: final polish and test verification"
```

---

## Phase 7: CI/CD with banua-coder-workflow

Uses reusable workflows from [banua-coder/banua-coder-workflow](https://github.com/banua-coder/banua-coder-workflow).

### Task 24: Go backend CI workflow

**Files:**
- Create: `.github/workflows/ci-backend.yml`

**Step 1: Create workflow file**
```yaml
# .github/workflows/ci-backend.yml
name: Backend CI

on:
  pull_request:
    paths:
      - 'backend/**'
      - '.github/workflows/ci-backend.yml'

jobs:
  ci:
    uses: banua-coder/banua-coder-workflow/.github/workflows/ci-go.yml@main
    with:
      go-version: '1.22'
      working-directory: backend
      run-tests: true
      run-lint: true
      run-build: true
      run-coverage: true
      test-packages: './...'
      build-target: './cmd/api'
      build-output: api
      coverage-threshold: 50
      post-coverage-comment: true
```

**Step 2: Commit**
```bash
git add .github/workflows/ci-backend.yml
git commit -m "ci: add Go backend CI using banua-coder-workflow"
```

---

### Task 25: Vue frontend CI workflow

**Files:**
- Create: `.github/workflows/ci-frontend.yml`

**Step 1: Create workflow file**
```yaml
# .github/workflows/ci-frontend.yml
name: Frontend CI

on:
  pull_request:
    paths:
      - 'frontend/**'
      - '.github/workflows/ci-frontend.yml'

jobs:
  ci:
    uses: banua-coder/banua-coder-workflow/.github/workflows/ci-node.yml@main
    with:
      node-version: '20'
      package-manager: npm
      working-directory: frontend
      run-lint: true
      run-typecheck: true
      run-tests: false  # Enable once vitest is configured
      run-build: true
      run-coverage: false
      post-coverage-comment: true
```

**Step 2: Commit**
```bash
git add .github/workflows/ci-frontend.yml
git commit -m "ci: add Vue frontend CI using banua-coder-workflow"
```

---

### Task 26: Release workflow (Git Flow)

**Files:**
- Create: `.github/workflows/release.yml`

**Step 1: Create workflow file**
```yaml
# .github/workflows/release.yml
name: Release

on:
  pull_request:
    types: [opened, synchronize, reopened, closed]
    branches: [main]
  push:
    tags:
      - 'v*'

jobs:
  release:
    uses: banua-coder/banua-coder-workflow/.github/workflows/release.yml@main
    with:
      project-type: web-app
      main-branch: main
      develop-branch: develop
      changelog-format: keepachangelog
      auto-merge-backport: true
    secrets:
      GH_PAT: ${{ secrets.GH_PAT }}
```

**Step 2: Commit**
```bash
git add .github/workflows/release.yml
git commit -m "ci: add release workflow with Git Flow support"
```

---

### Task 27: Deploy on tag workflow (SSH + Docker)

**Files:**
- Create: `.github/workflows/deploy.yml`
- Create: `scripts/deploy.sh`

**Step 1: Create deploy script**
```bash
#!/bin/bash
# scripts/deploy.sh
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
```

**Step 2: Create workflow file**
```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    tags:
      - 'v*'

jobs:
  deploy:
    uses: banua-coder/banua-coder-workflow/.github/workflows/deploy-on-tag.yml@main
    with:
      environment: production
      deploy-provider: ssh
      deploy-path: /opt/sulteng-procurement
      build-command: 'bash scripts/deploy.sh'
      backup-enabled: true
      backup-keep: 3
    secrets:
      SSH_HOST: ${{ secrets.SSH_HOST }}
      SSH_USER: ${{ secrets.SSH_USER }}
      SSH_KEY: ${{ secrets.SSH_KEY }}
```

**Step 3: Commit**
```bash
chmod +x scripts/deploy.sh
git add .github/workflows/deploy.yml scripts/deploy.sh
git commit -m "ci: add SSH deploy workflow with Docker build"
```

---

### Task 28: Housekeeping + sanity check workflows

**Files:**
- Create: `.github/workflows/housekeeping.yml`
- Create: `.github/workflows/sanity-check.yml`

**Step 1: Create housekeeping workflow**
```yaml
# .github/workflows/housekeeping.yml
name: Housekeeping

on:
  pull_request:
    types: [closed]

jobs:
  cleanup:
    if: github.event.pull_request.merged == true
    uses: banua-coder/banua-coder-workflow/.github/workflows/housekeeping.yml@main
    with:
      protected-branches: 'main,develop'
```

**Step 2: Create sanity check workflow**
```yaml
# .github/workflows/sanity-check.yml
name: Sanity Check

on:
  pull_request:
    paths:
      - 'frontend/**'

jobs:
  sanity:
    uses: banua-coder/banua-coder-workflow/.github/workflows/sanity-check.yml@main
    with:
      project-type: vue
      node-version: '20'
      vue-max-loc: 700
      check-vue-types: true
      post-comment: true
      fail-on-issues: false
```

**Step 3: Commit**
```bash
git add .github/workflows/housekeeping.yml .github/workflows/sanity-check.yml
git commit -m "ci: add housekeeping and sanity check workflows"
```

---

### Task 29: Daily scraper cron workflow

**Files:**
- Create: `.github/workflows/scraper-cron.yml`

**Step 1: Create scheduled scraper workflow**

This is a custom workflow (not from banua-coder-workflow) that triggers the scraper daily via SSH.

```yaml
# .github/workflows/scraper-cron.yml
name: Daily Scraper

on:
  schedule:
    - cron: '0 2 * * *'  # 2am UTC daily (9am WITA)
  workflow_dispatch:  # Allow manual trigger

jobs:
  scrape:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger scraper on server
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_KEY }}
          script: |
            cd /opt/sulteng-procurement
            docker compose run --rm scraper
            docker compose restart backend
            echo "Scraper completed at $(date)"
```

**Step 2: Commit**
```bash
git add .github/workflows/scraper-cron.yml
git commit -m "ci: add daily scraper cron workflow via SSH"
```

---

## Monorepo Structure (Updated)

```
sulteng-procurement/
├── .github/
│   └── workflows/
│       ├── ci-backend.yml       # Go CI (banua-coder-workflow)
│       ├── ci-frontend.yml      # Node CI (banua-coder-workflow)
│       ├── release.yml          # Git Flow release (banua-coder-workflow)
│       ├── deploy.yml           # SSH deploy on tag (banua-coder-workflow)
│       ├── housekeeping.yml     # Branch cleanup (banua-coder-workflow)
│       ├── sanity-check.yml     # Vue quality (banua-coder-workflow)
│       └── scraper-cron.yml     # Daily scraper (custom)
├── scripts/
│   └── deploy.sh               # Remote deploy script
├── docs/plans/
├── backend/
│   ├── cmd/
│   │   ├── api/main.go
│   │   └── scraper/main.go
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── domain/procurement.go
│   │   ├── scraper/sirup.go
│   │   ├── storage/parquet.go
│   │   ├── api/handler.go, router.go
│   │   └── service/procurement.go
│   ├── data/                    # Parquet files (gitignored)
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── App.vue
│   │   ├── api/procurement.ts
│   │   ├── composables/useProcurement.ts
│   │   ├── components/*.vue
│   │   └── types/procurement.ts
│   ├── package.json
│   ├── vite.config.ts
│   ├── nginx.conf
│   └── Dockerfile
├── docker-compose.yml
├── .env.example
└── .gitignore
```

---

## Task Summary

| # | Task | Phase | Est. |
|---|------|-------|------|
| 1 | Init monorepo + git | Scaffolding | 2min |
| 2 | Init Go backend module | Scaffolding | 3min |
| 3 | Define domain types | Scaffolding | 3min |
| 4 | SIRUP API client | Scraper | 5min |
| 5 | Parquet storage | Scraper | 5min |
| 6 | Scraper CLI | Scraper | 3min |
| 7 | Procurement service | API | 5min |
| 8 | HTTP handlers + router | API | 5min |
| 9 | API server main | API | 2min |
| 10 | Vue + Vite + Tailwind scaffold | Frontend | 3min |
| 11 | TS types + API client | Frontend | 3min |
| 12 | useProcurement composable | Frontend | 3min |
| 13 | FilterBar component | Frontend | 3min |
| 14 | SummaryCards component | Frontend | 2min |
| 15 | CategoryChart component | Frontend | 3min |
| 16 | TopProcurements component | Frontend | 2min |
| 17 | DataTable component | Frontend | 3min |
| 18 | App.vue dashboard | Frontend | 3min |
| 19 | Backend Dockerfile | Docker | 2min |
| 20 | Frontend Dockerfile | Docker | 2min |
| 21 | Docker Compose + Traefik | Docker | 3min |
| 22 | Cron daily scraper | Polish | 3min |
| 23 | Final test + verify | Polish | 3min |
| 24 | Go backend CI workflow | CI/CD | 3min |
| 25 | Vue frontend CI workflow | CI/CD | 3min |
| 26 | Release workflow (Git Flow) | CI/CD | 3min |
| 27 | Deploy on tag (SSH + Docker) | CI/CD | 5min |
| 28 | Housekeeping + sanity check | CI/CD | 3min |
| 29 | Daily scraper cron workflow | CI/CD | 3min |
