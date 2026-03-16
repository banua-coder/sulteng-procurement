# SPSE/SIPP Integration Plan — Tender Winners & Contract Realisasi

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enrich the existing SIRUP RUP (procurement plan) data with SPSE tender execution data — specifically the winning vendor, bid price, and contract value — so the dashboard can show actual contract utilisation against the planned budget.

**Architecture:** A new Go scraper (`cmd/spse-scraper`) fetches tender and non-tender data from the public SPSE portal at `https://spse.inaproc.id/sultengprov` using a CSRF-token extraction + DataTables POST pattern (same as the PyProc library). Results are stored in a separate `spse_{year}.parquet` file. The existing `ProcurementService` gains a `Join` method that links SIRUP records to SPSE results via `idReferensi` ↔ SPSE tender ID. A new `/api/v1/realisasi` endpoint surfaces joined records. The frontend adds a realisasi summary section and enriches the data table with winner and contract columns.

**Tech Stack:** Go 1.22, xitongsys/parquet-go, net/http (CSRF + DataTables POST), Vue 3 + Composition API, shadcn-vue, vue-chartjs

**Key linkage:** Every SIRUP record has an `idReferensi` field (the SPSE tender/package ID once the package moves to execution). This is the primary join key. Non-matched records (RUP not yet executed) are shown as "Belum ditenderkan."

**SPSE data access:** Public, no login needed. The portal embeds an `authenticityToken` in the HTML of the list pages. We extract it with a GET, then POST to `/dt/lelang` and `/dt/pl` with DataTables parameters. Rate-limit to 1 req/sec to be respectful.

---

## Phase 1: SPSE Domain Types & Storage

### Task 1: Define TenderResult and Winner domain types

**Files:**
- Create: `backend/internal/domain/tender.go`

**Step 1: Create the file**

```go
package domain

// TenderResult holds the execution outcome for one SPSE tender or pengadaan langsung package.
type TenderResult struct {
	ID          int64   `json:"id"          parquet:"name=id,           type=INT64"`
	NamaPaket   string  `json:"namaPaket"   parquet:"name=nama_paket,    type=BYTE_ARRAY, convertedtype=UTF8"`
	NilaiPagu   float64 `json:"nilaiPagu"   parquet:"name=nilai_pagu,    type=DOUBLE"`
	NilaiHPS    float64 `json:"nilaiHPS"    parquet:"name=nilai_hps,     type=DOUBLE"`
	NilaiKontrak float64 `json:"nilaiKontrak" parquet:"name=nilai_kontrak, type=DOUBLE"`
	Tahap       string  `json:"tahap"       parquet:"name=tahap,         type=BYTE_ARRAY, convertedtype=UTF8"`
	SatuanKerja string  `json:"satuanKerja" parquet:"name=satuan_kerja,  type=BYTE_ARRAY, convertedtype=UTF8"`
	Jenis       string  `json:"jenis"       parquet:"name=jenis,         type=BYTE_ARRAY, convertedtype=UTF8"` // "lelang" | "pl"
	Pemenang    string  `json:"pemenang"    parquet:"name=pemenang,      type=BYTE_ARRAY, convertedtype=UTF8"`
	NilaiPenawaran float64 `json:"nilaiPenawaran" parquet:"name=nilai_penawaran, type=DOUBLE"`
	NPWP        string  `json:"npwp"        parquet:"name=npwp,          type=BYTE_ARRAY, convertedtype=UTF8"`
}
```

**Step 2: Commit**
```bash
git add backend/internal/domain/tender.go
git commit -m "feat(domain): add TenderResult type for SPSE execution data"
```

---

### Task 2: Parquet storage for SPSE data

**Files:**
- Create: `backend/internal/storage/spse.go`
- Create: `backend/internal/storage/spse_test.go`

**Step 1: Write the failing test**

```go
// backend/internal/storage/spse_test.go
package storage

import (
	"os"
	"testing"
	"github.com/banua-coder/sulteng-procurement/backend/internal/domain"
)

func TestWriteAndReadTender(t *testing.T) {
	dir := t.TempDir()
	store := NewSpseStore(dir)

	input := []domain.TenderResult{
		{ID: 1, NamaPaket: "Paket A", NilaiKontrak: 500000000, Pemenang: "PT Contoh", Jenis: "lelang"},
		{ID: 2, NamaPaket: "Paket B", NilaiKontrak: 200000000, Pemenang: "CV Maju",   Jenis: "pl"},
	}

	if err := store.Write(2026, input); err != nil {
		t.Fatalf("Write: %v", err)
	}
	got, err := store.Read(2026)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 records, got %d", len(got))
	}
	if got[0].Pemenang != "PT Contoh" {
		t.Errorf("expected Pemenang='PT Contoh', got %q", got[0].Pemenang)
	}
}
```

**Step 2: Run test — expect FAIL**
```bash
cd backend && go test ./internal/storage/ -run TestWriteAndReadTender -v
# Expected: FAIL — NewSpseStore undefined
```

**Step 3: Implement `SpseStore`**

```go
// backend/internal/storage/spse.go
package storage

import (
	"fmt"
	"os"
	"path/filepath"

	goparquet "github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/writer"
	"github.com/xitongsys/parquet-go-source/local"

	"github.com/banua-coder/sulteng-procurement/backend/internal/domain"
)

// SpseStore persists SPSE tender results as parquet files.
type SpseStore struct {
	dir string
}

func NewSpseStore(dir string) *SpseStore {
	return &SpseStore{dir: dir}
}

func (s *SpseStore) path(year int) string {
	return filepath.Join(s.dir, fmt.Sprintf("spse_%d.parquet", year))
}

func (s *SpseStore) Exists(year int) bool {
	_, err := os.Stat(s.path(year))
	return err == nil
}

func (s *SpseStore) Write(year int, records []domain.TenderResult) error {
	fw, err := local.NewLocalFileWriter(s.path(year))
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	pw, err := writer.NewParquetWriter(fw, new(domain.TenderResult), 4)
	if err != nil {
		return fmt.Errorf("new writer: %w", err)
	}
	pw.CompressionType = goparquet.CompressionCodec_SNAPPY
	for _, r := range records {
		if err := pw.Write(r); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}
	if err := pw.WriteStop(); err != nil {
		return fmt.Errorf("write stop: %w", err)
	}
	return fw.Close()
}

func (s *SpseStore) Read(year int) ([]domain.TenderResult, error) {
	fr, err := local.NewLocalFileReader(s.path(year))
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer fr.Close()

	pr, err := reader.NewParquetReader(fr, new(domain.TenderResult), 4)
	if err != nil {
		return nil, fmt.Errorf("new reader: %w", err)
	}
	defer pr.ReadStop()

	n := int(pr.GetNumRows())
	records := make([]domain.TenderResult, n)
	if err := pr.Read(&records); err != nil {
		return nil, fmt.Errorf("read rows: %w", err)
	}
	return records, nil
}
```

**Step 4: Run test — expect PASS**
```bash
cd backend && go test ./internal/storage/ -run TestWriteAndReadTender -v
```

**Step 5: Commit**
```bash
git add backend/internal/storage/spse.go backend/internal/storage/spse_test.go
git commit -m "feat(storage): add SpseStore for SPSE tender results parquet"
```

---

## Phase 2: SPSE Scraper

### Task 3: SPSE HTTP client — token extraction

**Files:**
- Create: `backend/internal/scraper/spse.go`
- Create: `backend/internal/scraper/spse_test.go`

**Background:** The SPSE portal embeds `var authenticityToken = '...'` in the HTML of every list page. We need to extract this before making DataTables POST requests.

**Step 1: Write the failing test**

```go
// backend/internal/scraper/spse_test.go
package scraper

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtractToken(t *testing.T) {
	html := `<html><body><script>var authenticityToken = 'abc123xyz';</script></body></html>`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	defer ts.Close()

	client := NewSpseClient(ts.URL, 2026)
	token, err := client.extractToken("/")
	if err != nil {
		t.Fatalf("extractToken: %v", err)
	}
	if token != "abc123xyz" {
		t.Errorf("expected 'abc123xyz', got %q", token)
	}
}
```

**Step 2: Run test — expect FAIL**
```bash
cd backend && go test ./internal/scraper/ -run TestExtractToken -v
```

**Step 3: Implement `SpseClient` with `extractToken`**

```go
// backend/internal/scraper/spse.go
package scraper

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

var tokenRe = regexp.MustCompile(`authenticityToken\s*=\s*'([^']+)'`)

// SpseClient fetches tender execution data from an SPSE portal instance.
type SpseClient struct {
	baseURL string
	year    int
	http    *http.Client
}

func NewSpseClient(baseURL string, year int) *SpseClient {
	return &SpseClient{
		baseURL: baseURL,
		year:    year,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *SpseClient) extractToken(path string) (string, error) {
	resp, err := c.http.Get(c.baseURL + path)
	if err != nil {
		return "", fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	m := tokenRe.FindSubmatch(body)
	if m == nil {
		return "", fmt.Errorf("authenticityToken not found in %s", path)
	}
	return string(m[1]), nil
}
```

**Step 4: Run test — expect PASS**
```bash
cd backend && go test ./internal/scraper/ -run TestExtractToken -v
```

**Step 5: Commit**
```bash
git add backend/internal/scraper/spse.go backend/internal/scraper/spse_test.go
git commit -m "feat(scraper): add SpseClient with CSRF token extraction"
```

---

### Task 4: SPSE client — DataTables POST fetch

**Files:**
- Modify: `backend/internal/scraper/spse.go`
- Modify: `backend/internal/scraper/spse_test.go`

**Background:** After extracting the token, POST to `/dt/lelang` or `/dt/pl` with DataTables params. The response is a JSON object with `data` array, where each row is an array. Key indices for `lelang`:
- `[0]` = tender ID
- `[1]` = package name
- `[3]` = tahap (phase/status)
- `[5]` = satuan kerja name
- `[6]` = jenis (category)
- `[10]` = nilai kontrak

For `pl` (pengadaan langsung):
- `[0]` = package ID
- `[1]` = name
- `[3]` = tahap
- `[5]` = satuan kerja
- `[8]` = nilai kontrak

**Step 1: Add test for DataTables parsing**

```go
// add to spse_test.go
func TestParseLelangRow(t *testing.T) {
	row := []any{"123456", "Pembangunan Jalan", nil, "Selesai", nil, "Dinas PU", "Pekerjaan Konstruksi", nil, nil, nil, "5000000000", nil, nil, nil, nil}
	result, err := parseLelangRow(row)
	if err != nil {
		t.Fatalf("parseLelangRow: %v", err)
	}
	if result.ID != 123456 {
		t.Errorf("expected ID=123456, got %d", result.ID)
	}
	if result.NilaiKontrak != 5000000000 {
		t.Errorf("expected NilaiKontrak=5000000000, got %f", result.NilaiKontrak)
	}
	if result.Tahap != "Selesai" {
		t.Errorf("expected Tahap='Selesai', got %q", result.Tahap)
	}
}
```

**Step 2: Run test — expect FAIL**

**Step 3: Implement fetch + parse**

Add to `spse.go`:

```go
import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
)

type dtResponse struct {
	Data [][]any `json:"data"`
	RecordsTotal int `json:"recordsTotal"`
}

func parseLelangRow(row []any) (domain.TenderResult, error) {
	get := func(i int) string {
		if i >= len(row) || row[i] == nil { return "" }
		s, _ := row[i].(string)
		return strings.TrimSpace(s)
	}
	idStr := get(0)
	id, _ := strconv.ParseInt(idStr, 10, 64)
	kontrakStr := get(10)
	kontrak, _ := strconv.ParseFloat(strings.ReplaceAll(kontrakStr, ",", ""), 64)
	return domain.TenderResult{
		ID:           id,
		NamaPaket:    get(1),
		Tahap:        get(3),
		SatuanKerja:  get(5),
		NilaiKontrak: kontrak,
		Jenis:        "lelang",
	}, nil
}

func parsePLRow(row []any) (domain.TenderResult, error) {
	get := func(i int) string {
		if i >= len(row) || row[i] == nil { return "" }
		s, _ := row[i].(string)
		return strings.TrimSpace(s)
	}
	idStr := get(0)
	id, _ := strconv.ParseInt(idStr, 10, 64)
	kontrakStr := get(8)
	kontrak, _ := strconv.ParseFloat(strings.ReplaceAll(kontrakStr, ",", ""), 64)
	return domain.TenderResult{
		ID:           id,
		NamaPaket:    get(1),
		Tahap:        get(3),
		SatuanKerja:  get(5),
		NilaiKontrak: kontrak,
		Jenis:        "pl",
	}, nil
}

// fetchPage posts one DataTables page to the given endpoint path.
func (c *SpseClient) fetchPage(path, token string, start, length int) (*dtResponse, error) {
	form := url.Values{
		"authenticityToken": {token},
		"draw":              {"1"},
		"start":             {strconv.Itoa(start)},
		"length":            {strconv.Itoa(length)},
		"tahun":             {strconv.Itoa(c.year)},
	}
	resp, err := c.http.PostForm(c.baseURL+path, form)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", path, err)
	}
	defer resp.Body.Close()
	var result dtResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return &result, nil
}
```

**Step 4: Run test — expect PASS**
```bash
cd backend && go test ./internal/scraper/ -run TestParseLelangRow -v
```

**Step 5: Commit**
```bash
git add backend/internal/scraper/spse.go backend/internal/scraper/spse_test.go
git commit -m "feat(scraper): add DataTables POST fetch and row parsing for SPSE"
```

---

### Task 5: SPSE client — FetchAll (pagination)

**Files:**
- Modify: `backend/internal/scraper/spse.go`

**Step 1: Implement `FetchAll`**

```go
// FetchAll fetches all lelang and nontender pages for the configured year.
func (c *SpseClient) FetchAll() ([]domain.TenderResult, error) {
	const pageSize = 100
	var all []domain.TenderResult

	type endpoint struct {
		path  string
		parse func([]any) (domain.TenderResult, error)
		token string
	}

	tokenLelang, err := c.extractToken("/lelang")
	if err != nil {
		return nil, fmt.Errorf("token lelang: %w", err)
	}
	tokenPL, err := c.extractToken("/nontender")
	if err != nil {
		return nil, fmt.Errorf("token pl: %w", err)
	}

	endpoints := []endpoint{
		{"/dt/lelang", parseLelangRow, tokenLelang},
		{"/dt/pl",     parsePLRow,     tokenPL},
	}

	for _, ep := range endpoints {
		first, err := c.fetchPage(ep.path, ep.token, 0, pageSize)
		if err != nil {
			return nil, err
		}
		total := first.RecordsTotal
		log.Printf("SPSE %s: %d total records", ep.path, total)

		rows := first.Data
		for start := pageSize; start < total; start += pageSize {
			page, err := c.fetchPage(ep.path, ep.token, start, pageSize)
			if err != nil {
				return nil, err
			}
			rows = append(rows, page.Data...)
			time.Sleep(200 * time.Millisecond)
		}

		for _, row := range rows {
			r, err := ep.parse(row)
			if err != nil {
				continue // skip malformed rows
			}
			all = append(all, r)
		}
	}

	return all, nil
}
```

**Step 2: Build check**
```bash
cd backend && go build ./...
# Expected: success
```

**Step 3: Commit**
```bash
git add backend/internal/scraper/spse.go
git commit -m "feat(scraper): add FetchAll pagination for lelang and nontender"
```

---

### Task 6: SPSE winner enrichment — pemenang endpoint

**Files:**
- Modify: `backend/internal/scraper/spse.go`
- Modify: `backend/internal/scraper/spse_test.go`

**Background:** For completed tenders (`Tahap == "Selesai"` or contains "Pemenang"), fetch winner details from `/evaluasilel/{id}/pemenang` (lelang) or `/evaluasinontender/{id}/pemenang` (pl). Response is a JSON array of participants; filter for the winner flag.

**Step 1: Add test**

```go
func TestParseWinner(t *testing.T) {
	// Minimal winner JSON (array of participants)
	body := `[{"nama_peserta":"PT Maju Bersama","npwp":"12.345.678.9-000.000","nilai_penawaran":"4500000000","is_winner":"1"}]`
	winner, found := parseWinner([]byte(body))
	if !found {
		t.Fatal("expected winner found")
	}
	if winner.Pemenang != "PT Maju Bersama" {
		t.Errorf("expected 'PT Maju Bersama', got %q", winner.Pemenang)
	}
}
```

**Step 2: Run — expect FAIL**

**Step 3: Implement**

```go
type participantRec struct {
	NamaPeserta    string `json:"nama_peserta"`
	NPWP           string `json:"npwp"`
	NilaiPenawaran string `json:"nilai_penawaran"`
	IsWinner       string `json:"is_winner"`
}

func parseWinner(body []byte) (domain.TenderResult, bool) {
	var participants []participantRec
	if err := json.Unmarshal(body, &participants); err != nil {
		return domain.TenderResult{}, false
	}
	for _, p := range participants {
		if p.IsWinner == "1" {
			val, _ := strconv.ParseFloat(strings.ReplaceAll(p.NilaiPenawaran, ",", ""), 64)
			return domain.TenderResult{
				Pemenang:       p.NamaPeserta,
				NPWP:           p.NPWP,
				NilaiPenawaran: val,
			}, true
		}
	}
	return domain.TenderResult{}, false
}

// EnrichWinners fetches pemenang data for completed records and merges it in.
func (c *SpseClient) EnrichWinners(records []domain.TenderResult) []domain.TenderResult {
	enriched := make([]domain.TenderResult, len(records))
	for i, r := range records {
		if !strings.Contains(strings.ToLower(r.Tahap), "selesai") {
			enriched[i] = r
			continue
		}
		path := fmt.Sprintf("/evaluasilel/%d/pemenang", r.ID)
		if r.Jenis == "pl" {
			path = fmt.Sprintf("/evaluasinontender/%d/pemenang", r.ID)
		}
		resp, err := c.http.Get(c.baseURL + path)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if w, ok := parseWinner(body); ok {
				r.Pemenang = w.Pemenang
				r.NPWP = w.NPWP
				r.NilaiPenawaran = w.NilaiPenawaran
			}
		}
		enriched[i] = r
		time.Sleep(100 * time.Millisecond)
	}
	return enriched
}
```

**Step 4: Run test — expect PASS**
```bash
cd backend && go test ./internal/scraper/ -run TestParseWinner -v
```

**Step 5: Commit**
```bash
git add backend/internal/scraper/spse.go backend/internal/scraper/spse_test.go
git commit -m "feat(scraper): add winner enrichment from pemenang endpoint"
```

---

### Task 7: SPSE scraper CLI

**Files:**
- Create: `backend/cmd/spse-scraper/main.go`

**Step 1: Create**

```go
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
```

**Step 2: Build check**
```bash
cd backend && go build ./cmd/spse-scraper
# Expected: success
```

**Step 3: Add to Makefile** (add after `scrape:` target)

```makefile
scrape-spse: ## Fetch SPSE tender winner data and write to parquet
	@mkdir -p $(DATA_DIR)
	cd backend && DATA_DIR=../$(DATA_DIR) SCRAPER_YEAR=$(SCRAPER_YEAR) go run ./cmd/spse-scraper
```

**Step 4: Add to docker-compose.yml** — add a new service under `scraper`:

```yaml
  spse-scraper:
    build: ./backend
    command: ["/app/spse-scraper"]
    environment:
      - DATA_DIR=/app/data
      - SCRAPER_YEAR=${SCRAPER_YEAR:-2026}
      - SPSE_URL=${SPSE_URL:-https://spse.inaproc.id/sultengprov}
    volumes:
      - procurement-data:/app/data
    profiles:
      - scrape
```

**Step 5: Update backend Dockerfile** to build the new binary.

Current `Dockerfile` `RUN` line:
```dockerfile
RUN go build -o /app/api ./cmd/api && \
    go build -o /app/scraper ./cmd/scraper
```

Change to:
```dockerfile
RUN go build -o /app/api ./cmd/api && \
    go build -o /app/scraper ./cmd/scraper && \
    go build -o /app/spse-scraper ./cmd/spse-scraper
```

**Step 6: Add `SPSE_URL` to `.env.example`**
```
SPSE_URL=https://spse.inaproc.id/sultengprov
```

**Step 7: Commit**
```bash
git add backend/cmd/spse-scraper/ Makefile docker-compose.yml backend/Dockerfile .env.example
git commit -m "feat: add spse-scraper CLI, Makefile target, and Docker service"
```

---

## Phase 3: Service — Join RUP with SPSE Data

### Task 8: JoinedRecord domain type

**Files:**
- Create: `backend/internal/domain/realisasi.go`

**Step 1: Create**

```go
package domain

// JoinedRecord links a SIRUP RUP plan record with its SPSE execution result.
// If the package has not yet been tendered, TenderResult is nil.
type JoinedRecord struct {
	RUP    Procurement   `json:"rup"`
	Tender *TenderResult `json:"tender,omitempty"`
}

// RealisasiSummary provides aggregate budget utilisation metrics.
type RealisasiSummary struct {
	TotalPagu      float64 `json:"totalPagu"`
	TotalKontrak   float64 `json:"totalKontrak"`
	TotalSelesai   int     `json:"totalSelesai"`
	UtilisasiRate  float64 `json:"utilisasiRate"` // totalKontrak / totalPagu * 100
	BelumTender    int     `json:"belumTender"`
}
```

**Step 2: Commit**
```bash
git add backend/internal/domain/realisasi.go
git commit -m "feat(domain): add JoinedRecord and RealisasiSummary types"
```

---

### Task 9: RealisasiService

**Files:**
- Create: `backend/internal/service/realisasi.go`
- Create: `backend/internal/service/realisasi_test.go`

**Step 1: Write failing tests**

```go
// backend/internal/service/realisasi_test.go
package service

import (
	"testing"
	"github.com/banua-coder/sulteng-procurement/backend/internal/domain"
)

func TestJoin(t *testing.T) {
	rup := []domain.Procurement{
		{ID: 1, IdReferensi: 100, Paket: "Jalan Raya", Pagu: 5000000000},
		{ID: 2, IdReferensi: 0,   Paket: "Komputer",   Pagu: 100000000},
	}
	tenders := []domain.TenderResult{
		{ID: 100, NilaiKontrak: 4800000000, Pemenang: "PT Jalan Raya"},
	}

	svc := NewRealisasiService(rup, tenders)
	joined := svc.Join()

	if len(joined) != 2 {
		t.Fatalf("expected 2 joined records, got %d", len(joined))
	}
	// First record should be matched
	if joined[0].Tender == nil {
		t.Error("expected Tender to be non-nil for matched record")
	}
	if joined[0].Tender.Pemenang != "PT Jalan Raya" {
		t.Errorf("expected Pemenang='PT Jalan Raya', got %q", joined[0].Tender.Pemenang)
	}
	// Second record has no match
	if joined[1].Tender != nil {
		t.Error("expected Tender to be nil for unmatched record")
	}
}

func TestRealisasiSummary(t *testing.T) {
	rup := []domain.Procurement{
		{ID: 1, IdReferensi: 100, Pagu: 5000000000},
		{ID: 2, IdReferensi: 0,   Pagu: 1000000000},
	}
	tenders := []domain.TenderResult{
		{ID: 100, NilaiKontrak: 4800000000, Tahap: "Selesai"},
	}

	svc := NewRealisasiService(rup, tenders)
	summary := svc.GetSummary()

	if summary.TotalKontrak != 4800000000 {
		t.Errorf("expected TotalKontrak=4800000000, got %f", summary.TotalKontrak)
	}
	if summary.TotalSelesai != 1 {
		t.Errorf("expected TotalSelesai=1, got %d", summary.TotalSelesai)
	}
	if summary.BelumTender != 1 {
		t.Errorf("expected BelumTender=1, got %d", summary.BelumTender)
	}
}
```

**Step 2: Run — expect FAIL**
```bash
cd backend && go test ./internal/service/ -run TestJoin -v
```

**Step 3: Implement**

```go
// backend/internal/service/realisasi.go
package service

import (
	"strings"
	"github.com/banua-coder/sulteng-procurement/backend/internal/domain"
)

// RealisasiService joins RUP plan data with SPSE tender execution data.
type RealisasiService struct {
	rup     []domain.Procurement
	tenders []domain.TenderResult
	index   map[int64]*domain.TenderResult // keyed by SPSE tender ID
}

func NewRealisasiService(rup []domain.Procurement, tenders []domain.TenderResult) *RealisasiService {
	idx := make(map[int64]*domain.TenderResult, len(tenders))
	for i := range tenders {
		idx[tenders[i].ID] = &tenders[i]
	}
	return &RealisasiService{rup: rup, tenders: tenders, index: idx}
}

// Join matches each RUP record to its SPSE result by IdReferensi.
func (s *RealisasiService) Join() []domain.JoinedRecord {
	result := make([]domain.JoinedRecord, len(s.rup))
	for i, r := range s.rup {
		jr := domain.JoinedRecord{RUP: r}
		if r.IdReferensi != 0 {
			if t, ok := s.index[int64(r.IdReferensi)]; ok {
				jr.Tender = t
			}
		}
		result[i] = jr
	}
	return result
}

// GetSummary computes aggregate utilisation metrics across all joined records.
func (s *RealisasiService) GetSummary() domain.RealisasiSummary {
	var totalPagu, totalKontrak float64
	var selesai, belum int

	for _, r := range s.rup {
		totalPagu += r.Pagu
		if t, ok := s.index[int64(r.IdReferensi)]; ok && r.IdReferensi != 0 {
			totalKontrak += t.NilaiKontrak
			if strings.Contains(strings.ToLower(t.Tahap), "selesai") {
				selesai++
			}
		} else {
			belum++
		}
	}

	var rate float64
	if totalPagu > 0 {
		rate = totalKontrak / totalPagu * 100
	}

	return domain.RealisasiSummary{
		TotalPagu:     totalPagu,
		TotalKontrak:  totalKontrak,
		TotalSelesai:  selesai,
		UtilisasiRate: rate,
		BelumTender:   belum,
	}
}
```

**Step 4: Run tests — expect PASS**
```bash
cd backend && go test ./internal/service/ -run "TestJoin|TestRealisasiSummary" -v
```

**Step 5: Commit**
```bash
git add backend/internal/service/realisasi.go backend/internal/service/realisasi_test.go
git commit -m "feat(service): add RealisasiService joining RUP with SPSE tender data"
```

---

## Phase 4: API Endpoints

### Task 10: Realisasi handler + router integration

**Files:**
- Modify: `backend/internal/api/handler.go`
- Modify: `backend/internal/api/router.go`
- Modify: `backend/cmd/api/main.go`

**Step 1: Add handler tests**

```go
// add to handler_test.go
func TestGetRealisasiSummary(t *testing.T) {
	rup := []domain.Procurement{
		{ID: 1, IdReferensi: 10, Paket: "Jalan", Pagu: 5000000000},
	}
	tenders := []domain.TenderResult{
		{ID: 10, NilaiKontrak: 4500000000, Tahap: "Selesai", Pemenang: "PT Test"},
	}
	svc := service.NewProcurementService(rup)
	realSvc := service.NewRealisasiService(rup, tenders)
	h := NewHandler(svc)
	h.SetRealisasiService(realSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/realisasi/summary", nil)
	w := httptest.NewRecorder()
	h.GetRealisasiSummary(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var result domain.RealisasiSummary
	json.NewDecoder(w.Body).Decode(&result)
	if result.TotalSelesai != 1 {
		t.Errorf("expected 1 selesai, got %d", result.TotalSelesai)
	}
}
```

**Step 2: Run — expect FAIL**

**Step 3: Implement**

Add to `handler.go`:

```go
// in Handler struct
type Handler struct {
	svc      *service.ProcurementService
	realSvc  *service.RealisasiService
}

func (h *Handler) SetRealisasiService(svc *service.RealisasiService) {
	h.realSvc = svc
}

func (h *Handler) GetRealisasiSummary(w http.ResponseWriter, r *http.Request) {
	if h.realSvc == nil {
		http.Error(w, "SPSE data not loaded", http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, h.realSvc.GetSummary())
}

func (h *Handler) GetRealisasi(w http.ResponseWriter, r *http.Request) {
	if h.realSvc == nil {
		http.Error(w, "SPSE data not loaded", http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, h.realSvc.Join())
}
```

Add to `router.go` under `/api/v1`:
```go
r.Get("/realisasi/summary", h.GetRealisasiSummary)
r.Get("/realisasi",         h.GetRealisasi)
```

Update `cmd/api/main.go` to load SPSE parquet if it exists and wire `RealisasiService`:
```go
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
```

**Step 4: Run all tests — expect PASS**
```bash
cd backend && go test ./... -v
```

**Step 5: Commit**
```bash
git add backend/internal/api/ backend/cmd/api/main.go
git commit -m "feat(api): add /realisasi and /realisasi/summary endpoints"
```

---

## Phase 5: Frontend

### Task 11: Add TypeScript types for realisasi

**Files:**
- Modify: `frontend/src/types/procurement.ts`

**Step 1: Add types**

```typescript
export interface TenderResult {
  id: number
  namaPaket: string
  nilaiKontrak: number
  nilaiPenawaran: number
  tahap: string
  satuanKerja: string
  jenis: 'lelang' | 'pl'
  pemenang: string
  npwp: string
}

export interface JoinedRecord {
  rup: Procurement
  tender?: TenderResult
}

export interface RealisasiSummary {
  totalPagu: number
  totalKontrak: number
  totalSelesai: number
  utilisasiRate: number
  belumTender: number
}
```

**Step 2: Commit**
```bash
git add frontend/src/types/procurement.ts
git commit -m "feat(types): add TenderResult, JoinedRecord, RealisasiSummary"
```

---

### Task 12: Realisasi API client + composable

**Files:**
- Modify: `frontend/src/api/procurement.ts`
- Create: `frontend/src/composables/useRealisasi.ts`

**Step 1: Add API functions to `procurement.ts`**

```typescript
export function getRealisasiSummary(): Promise<RealisasiSummary> {
  return fetchJSON(`${BASE}/realisasi/summary`)
}

export function getRealisasi(): Promise<JoinedRecord[]> {
  return fetchJSON(`${BASE}/realisasi`)
}
```

**Step 2: Create composable**

```typescript
// frontend/src/composables/useRealisasi.ts
import { ref, onMounted } from 'vue'
import { getRealisasiSummary, getRealisasi } from '../api/procurement'
import type { RealisasiSummary, JoinedRecord } from '../types/procurement'

export function useRealisasi() {
  const summary = ref<RealisasiSummary | null>(null)
  const records = ref<JoinedRecord[]>([])
  const loading = ref(false)
  const available = ref(false) // false when API returns 503

  async function load() {
    loading.value = true
    try {
      summary.value = await getRealisasiSummary()
      records.value = await getRealisasi()
      available.value = true
    } catch {
      available.value = false
    } finally {
      loading.value = false
    }
  }

  onMounted(load)

  return { summary, records, loading, available }
}
```

**Step 3: Commit**
```bash
git add frontend/src/api/procurement.ts frontend/src/composables/useRealisasi.ts
git commit -m "feat(frontend): add realisasi API client and useRealisasi composable"
```

---

### Task 13: RealisasiCards summary component

**Files:**
- Create: `frontend/src/components/RealisasiCards.vue`

**Step 1: Create component**

```vue
<script setup lang="ts">
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import type { RealisasiSummary } from '../types/procurement'
import { formatRupiah } from '@/utils/format'

defineProps<{ summary: RealisasiSummary }>()
</script>

<template>
  <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
    <Card>
      <CardHeader class="pb-1">
        <CardTitle class="text-sm font-medium text-muted-foreground">Total kontrak</CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-2xl font-bold">{{ formatRupiah(summary.totalKontrak) }}</p>
        <p class="text-xs text-muted-foreground mt-1">dari {{ formatRupiah(summary.totalPagu) }} pagu</p>
      </CardContent>
    </Card>

    <Card>
      <CardHeader class="pb-1">
        <CardTitle class="text-sm font-medium text-muted-foreground">Utilisasi anggaran</CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-2xl font-bold">{{ summary.utilisasiRate.toFixed(1) }}%</p>
        <div class="w-full bg-stone-200 rounded-full h-1.5 mt-2">
          <div class="bg-stone-700 h-1.5 rounded-full" :style="{ width: Math.min(summary.utilisasiRate, 100) + '%' }" />
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader class="pb-1">
        <CardTitle class="text-sm font-medium text-muted-foreground">Paket selesai</CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-2xl font-bold">{{ summary.totalSelesai.toLocaleString('id-ID') }}</p>
        <Badge variant="secondary" class="mt-1 text-xs">Kontrak ditandatangani</Badge>
      </CardContent>
    </Card>

    <Card>
      <CardHeader class="pb-1">
        <CardTitle class="text-sm font-medium text-muted-foreground">Belum ditenderkan</CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-2xl font-bold">{{ summary.belumTender.toLocaleString('id-ID') }}</p>
        <p class="text-xs text-muted-foreground mt-1">Masih di tahap perencanaan</p>
      </CardContent>
    </Card>
  </div>
</template>
```

**Step 2: Commit**
```bash
git add frontend/src/components/RealisasiCards.vue
git commit -m "feat(ui): add RealisasiCards summary component"
```

---

### Task 14: Enrich DataTable with pemenang + nilai kontrak columns

**Files:**
- Modify: `frontend/src/components/DataTable.vue`

**Background:** The existing `DataTable` receives `PaginatedResult`. Change the `/api/v1/procurements` backend response to optionally include SPSE fields by adding a `WithTender` flag to the query — OR keep it simple and show pemenang/kontrak only in the `DataTable` when the data is available via `/api/v1/realisasi`. The simplest approach is to add two optional columns to the existing table that display when realisasi data is passed as an optional prop.

**Step 1: Update DataTable props**

```typescript
// in DataTable.vue <script setup>
import type { JoinedRecord, PaginatedResult } from '../types/procurement'

const props = defineProps<{
  result: PaginatedResult | null
  loading: boolean
  sortBy: string
  sortDir: string
  // Optional: map from RUP id → TenderResult for enrichment
  tenderMap?: Map<number, import('../types/procurement').TenderResult>
}>()
```

**Step 2: Add columns to table header and rows**

In the table header, after the existing `Pagu` column:
```html
<TableHead v-if="tenderMap">Nilai Kontrak</TableHead>
<TableHead v-if="tenderMap">Pemenang</TableHead>
```

In the table row, after `Pagu`:
```html
<TableCell v-if="tenderMap">
  {{ tenderMap.get(item.id) ? formatRupiah(tenderMap.get(item.id)!.nilaiKontrak) : '—' }}
</TableCell>
<TableCell v-if="tenderMap" class="text-xs max-w-32 truncate">
  {{ tenderMap.get(item.id)?.pemenang ?? '—' }}
</TableCell>
```

**Step 3: Update App.vue to build tenderMap from useRealisasi records**

```typescript
// in App.vue
import { useRealisasi } from './composables/useRealisasi'
const { summary: realSummary, records: realRecords, available: spseAvailable } = useRealisasi()

const tenderMap = computed(() => {
  if (!spseAvailable.value) return undefined
  const map = new Map<number, TenderResult>()
  for (const r of realRecords.value) {
    if (r.tender) map.set(r.rup.id, r.tender)
  }
  return map
})
```

Pass to `DataTable`:
```html
<DataTable :tender-map="tenderMap" ... />
```

Also add `RealisasiCards` to App.vue below the existing summary cards, conditionally:
```html
<div v-if="spseAvailable && realSummary" class="mt-4">
  <h2 class="text-lg font-semibold text-stone-700 mb-3">Realisasi kontrak</h2>
  <RealisasiCards :summary="realSummary" />
</div>
```

**Step 4: Commit**
```bash
git add frontend/src/components/DataTable.vue frontend/src/App.vue
git commit -m "feat(ui): enrich DataTable with pemenang and nilai kontrak when SPSE available"
```

---

## Phase 6: Cron + Deployment

### Task 15: Wire SPSE scraper into daily cron + update CI

**Files:**
- Modify: `backend/cmd/api/main.go`
- Modify: `.github/workflows/scraper-cron.yml`

**Step 1: Add SPSE scraper to daily cron in main.go**

After the existing SIRUP `runScrape` function, add:

```go
runSpseScrape := func() {
    log.Println("Cron: starting SPSE scrape")
    spseBase := os.Getenv("SPSE_URL")
    if spseBase == "" {
        spseBase = "https://spse.inaproc.id/sultengprov"
    }
    spseClient := scraper.NewSpseClient(spseBase, cfg.ScraperYear)
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
    mu.Lock()
    realSvc := service.NewRealisasiService(/* current data */ nil, tenders)
    handler.SetRealisasiService(realSvc)
    mu.Unlock()
    log.Printf("Cron: SPSE scrape complete, %d tenders loaded", len(tenders))
}

c.AddFunc(cfg.CronSchedule, runSpseScrape)
```

**Step 2: Update scraper-cron.yml** to also run spse-scraper:

```yaml
script: |
  cd /opt/sulteng-procurement
  docker compose run --rm scraper
  docker compose run --rm spse-scraper
  docker compose restart backend
  echo "All scrapers completed at $(date)"
```

**Step 3: Run all tests**
```bash
cd backend && go test ./... -v
# Expected: all pass
```

**Step 4: Build all binaries**
```bash
cd backend && go build ./...
# Expected: success
```

**Step 5: Commit**
```bash
git add backend/cmd/api/main.go .github/workflows/scraper-cron.yml
git commit -m "feat: wire SPSE scraper into daily cron and CI workflow"
```

---

## Monorepo Changes Summary

```
backend/
├── cmd/spse-scraper/main.go       # NEW
├── internal/domain/tender.go      # NEW: TenderResult
├── internal/domain/realisasi.go   # NEW: JoinedRecord, RealisasiSummary
├── internal/scraper/spse.go       # NEW: SpseClient
├── internal/storage/spse.go       # NEW: SpseStore
├── internal/service/realisasi.go  # NEW: RealisasiService
├── internal/api/handler.go        # MODIFIED: SetRealisasiService, GetRealisasi*
├── internal/api/router.go         # MODIFIED: /realisasi routes
└── cmd/api/main.go                # MODIFIED: load SPSE + cron

frontend/src/
├── types/procurement.ts           # MODIFIED: TenderResult, JoinedRecord, RealisasiSummary
├── api/procurement.ts             # MODIFIED: getRealisasi*, getRealisasiSummary
├── composables/useRealisasi.ts    # NEW
├── components/RealisasiCards.vue  # NEW
├── components/DataTable.vue       # MODIFIED: optional pemenang/kontrak columns
└── App.vue                        # MODIFIED: useRealisasi, tenderMap, RealisasiCards
```

## Required Secrets (already in repo)
- `SSH_HOST`, `SSH_USER`, `SSH_KEY` — for deployment
- No new secrets needed; SPSE URL is public

## Known Limitations
1. `IdReferensi` in SIRUP data is only populated once a package moves to the tendering phase — packages still in "perencanaan" will not match
2. SPSE `authenticityToken` is session-bound; if CSRF protection changes, the scraper needs updating
3. Winner enrichment makes one HTTP request per completed tender — for large datasets this is slow; consider batching or caching
4. SPSE may rate-limit aggressive scrapers; the 200ms delay between pages should be sufficient but monitor for 429s
