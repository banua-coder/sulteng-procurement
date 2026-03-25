package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/banua-coder/sulteng-procurement/backend/internal/domain"
)

const browserUA = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

// uaTransport injects a browser User-Agent on every outgoing request so the
// SPSE portal does not reject the Go default agent with 403.
type uaTransport struct{ base http.RoundTripper }

func (t *uaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("User-Agent", browserUA)
	return t.base.RoundTrip(clone)
}

var tokenRe = regexp.MustCompile(`authenticityToken\s*=\s*'([^']+)'`)

// SpseClient fetches tender execution data from an SPSE portal instance.
type SpseClient struct {
	baseURL string
	year    int
	http    *http.Client
}

// NewSpseClient creates a client targeting the given base URL and budget year.
// It uses a cookie jar so the session cookie from the initial page load is
// automatically sent on subsequent DataTables POST requests.
func NewSpseClient(baseURL string, year int) *SpseClient {
	jar, _ := cookiejar.New(nil)
	return &SpseClient{
		baseURL: baseURL,
		year:    year,
		http: &http.Client{
			Timeout:   30 * time.Second,
			Jar:       jar,
			Transport: &uaTransport{base: http.DefaultTransport},
		},
	}
}

func (c *SpseClient) extractToken(path string) (string, error) {
	resp, err := c.http.Get(c.baseURL + path)
	if err != nil {
		return "", fmt.Errorf("GET %s: %w", path, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = resp.Body.Close()
		return "", fmt.Errorf("GET %s: HTTP %d", path, resp.StatusCode)
	}
	defer func() { _ = resp.Body.Close() }()
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

type dtResponse struct {
	Data         [][]any `json:"data"`
	RecordsTotal int     `json:"recordsTotal"`
}

func parseLelangRow(row []any) (domain.TenderResult, error) {
	if len(row) == 0 {
		return domain.TenderResult{}, fmt.Errorf("empty row")
	}
	get := func(i int) string {
		if i >= len(row) || row[i] == nil {
			return ""
		}
		s, _ := row[i].(string)
		return strings.TrimSpace(s)
	}
	id, _ := strconv.ParseInt(get(0), 10, 64)
	kontrak, _ := strconv.ParseFloat(strings.ReplaceAll(get(10), ",", ""), 64)
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
	if len(row) == 0 {
		return domain.TenderResult{}, fmt.Errorf("empty row")
	}
	get := func(i int) string {
		if i >= len(row) || row[i] == nil {
			return ""
		}
		s, _ := row[i].(string)
		return strings.TrimSpace(s)
	}
	id, _ := strconv.ParseInt(get(0), 10, 64)
	kontrak, _ := strconv.ParseFloat(strings.ReplaceAll(get(8), ",", ""), 64)
	return domain.TenderResult{
		ID:           id,
		NamaPaket:    get(1),
		Tahap:        get(3),
		SatuanKerja:  get(5),
		NilaiKontrak: kontrak,
		Jenis:        "pl",
	}, nil
}

// fetchPage POSTs one DataTables page to the given endpoint path.
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("POST %s: HTTP %d", path, resp.StatusCode)
	}
	defer func() { _ = resp.Body.Close() }()
	var result dtResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return &result, nil
}

// FetchAll fetches all lelang and nontender pages for the configured year.
func (c *SpseClient) FetchAll() ([]domain.TenderResult, error) {
	const pageSize = 100
	var all []domain.TenderResult

	type endpoint struct {
		listPath string
		dtPath   string
		parse    func([]any) (domain.TenderResult, error)
	}

	endpoints := []endpoint{
		{"/lelang", "/dt/lelang", parseLelangRow},
		{"/nontender", "/dt/pl", parsePLRow},
	}

	for _, ep := range endpoints {
		token, err := c.extractToken(ep.listPath)
		if err != nil {
			return nil, fmt.Errorf("token %s: %w", ep.listPath, err)
		}

		first, err := c.fetchPage(ep.dtPath, token, 0, pageSize)
		if err != nil {
			return nil, err
		}
		total := first.RecordsTotal
		log.Printf("SPSE %s: %d total records", ep.dtPath, total)

		rows := first.Data
		for start := pageSize; start < total; start += pageSize {
			page, err := c.fetchPage(ep.dtPath, token, start, pageSize)
			if err != nil {
				return nil, err
			}
			rows = append(rows, page.Data...)
			time.Sleep(200 * time.Millisecond)
		}

		for _, row := range rows {
			r, err := ep.parse(row)
			if err != nil {
				continue
			}
			all = append(all, r)
		}
	}

	return all, nil
}

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
		if err != nil {
			log.Printf("EnrichWinners: GET %s: %v", path, err)
			enriched[i] = r
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()
			log.Printf("EnrichWinners: GET %s: HTTP %d", path, resp.StatusCode)
			enriched[i] = r
			time.Sleep(100 * time.Millisecond)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if w, ok := parseWinner(body); ok {
			r.Pemenang = w.Pemenang
			r.NPWP = w.NPWP
			r.NilaiPenawaran = w.NilaiPenawaran
		}
		enriched[i] = r
		time.Sleep(100 * time.Millisecond)
	}
	return enriched
}
