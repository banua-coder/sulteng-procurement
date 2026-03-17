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
		if r.IdReferensi != 0 {
			if t, ok := s.index[int64(r.IdReferensi)]; ok {
				totalKontrak += t.NilaiKontrak
				if strings.Contains(strings.ToLower(t.Tahap), "selesai") {
					selesai++
				}
				continue
			}
		}
		belum++
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
