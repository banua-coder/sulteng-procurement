package scraper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/banua-coder/sulteng-procurement/backend/internal/domain"
)

// SultengKLDICodes lists the 14 KLPD codes for Sulawesi Tengah.
var SultengKLDICodes = []int{569, 575, 570, 571, 578, 572, 187, 573, 579, 576, 574, 577, 782, 787}

// SirupRecord maps the raw JSON fields from the SIRUP search API.
type SirupRecord struct {
	ID               int64   `json:"id"`
	Paket            string  `json:"paket"`
	Pagu             float64 `json:"pagu"`
	JenisPengadaan   string  `json:"jenisPengadaan"`
	Metode           string  `json:"metode"`
	Pemilihan        string  `json:"pemilihan"`
	SatuanKerja      string  `json:"satuanKerja"`
	KLDI             string  `json:"kldi"`
	Lokasi           string  `json:"lokasi"`
	SumberDana       string  `json:"sumberDana"`
	IsPDN            bool    `json:"isPDN"`
	IsUMK            bool    `json:"isUMK"`
	IdBulan          int     `json:"idBulan"`
	IdKldi           string  `json:"idKldi"`
	IdReferensi      int     `json:"id_referensi"`
	IdSatker         int     `json:"idSatker"`
	IdMetode         int     `json:"idMetode"`
	IdJenisPengadaan int     `json:"idJenisPengadaan"`
}

// SirupResponse is the top-level SIRUP API response envelope.
type SirupResponse struct {
	RecordsTotal    int           `json:"recordsTotal"`
	RecordsFiltered int           `json:"recordsFiltered"`
	Data            []SirupRecord `json:"data"`
}

// SirupClient fetches procurement data from the SIRUP search API.
type SirupClient struct {
	baseURL string
	year    int
	http    *http.Client
}

// NewSirupClient creates a client targeting the given base URL and budget year.
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

// FetchPage retrieves one page of results starting at offset start with the given page length.
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

// FetchAll fetches all pages and returns every record for the configured year and KLDI codes.
func (c *SirupClient) FetchAll() ([]SirupRecord, error) {
	const pageSize = 100
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
		time.Sleep(200 * time.Millisecond)
	}

	return all, nil
}

// ToDomain converts raw SIRUP records to domain procurement types.
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
			IdReferensi:      int64(r.IdReferensi),
			IdSatker:         r.IdSatker,
			IdMetode:         r.IdMetode,
			IdJenisPengadaan: r.IdJenisPengadaan,
		}
	}
	return result
}
