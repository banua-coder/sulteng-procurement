package service

import (
	"testing"

	"github.com/banua-coder/sulteng-procurement/backend/internal/domain"
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
