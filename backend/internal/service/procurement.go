package service

import (
	"math"
	"sort"
	"strings"

	"github.com/banua-coder/sulteng-procurement/backend/internal/domain"
)

// ProcurementService holds in-memory procurement data and provides query and aggregation methods.
type ProcurementService struct {
	data []domain.Procurement
}

// NewProcurementService creates a service seeded with the given records.
func NewProcurementService(data []domain.Procurement) *ProcurementService {
	return &ProcurementService{data: data}
}

// SetData replaces the in-memory dataset (used by the cron refresh).
func (s *ProcurementService) SetData(data []domain.Procurement) {
	s.data = data
}

// GetSummary returns aggregated totals broken down by jenis, KLDI, and metode.
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

// Query filters, sorts, and paginates the in-memory dataset.
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

// GetFilters returns the distinct values for each filterable field, sorted alphabetically.
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
