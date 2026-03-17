package service

import (
	"testing"

	"github.com/banua-coder/sulteng-procurement/backend/internal/domain"
)

func TestJoin(t *testing.T) {
	rup := []domain.Procurement{
		{ID: 1, IdReferensi: 100, Paket: "Jalan Raya", Pagu: 5000000000},
		{ID: 2, IdReferensi: 0, Paket: "Komputer", Pagu: 100000000},
	}
	tenders := []domain.TenderResult{
		{ID: 100, NilaiKontrak: 4800000000, Pemenang: "PT Jalan Raya"},
	}

	svc := NewRealisasiService(rup, tenders)
	joined := svc.Join()

	if len(joined) != 2 {
		t.Fatalf("expected 2 joined records, got %d", len(joined))
	}
	if joined[0].Tender == nil {
		t.Error("expected Tender to be non-nil for matched record")
	}
	if joined[0].Tender.Pemenang != "PT Jalan Raya" {
		t.Errorf("expected Pemenang='PT Jalan Raya', got %q", joined[0].Tender.Pemenang)
	}
	if joined[1].Tender != nil {
		t.Error("expected Tender to be nil for unmatched record")
	}
}

func TestRealisasiSummary(t *testing.T) {
	rup := []domain.Procurement{
		{ID: 1, IdReferensi: 100, Pagu: 5000000000},
		{ID: 2, IdReferensi: 0, Pagu: 1000000000},
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
