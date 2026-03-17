package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/banua-coder/sulteng-procurement/backend/internal/domain"
	"github.com/banua-coder/sulteng-procurement/backend/internal/service"
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
	if err := json.NewDecoder(w.Body).Decode(&summary); err != nil {
		t.Fatalf("decode error: %v", err)
	}
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
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
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
	if err := json.NewDecoder(w.Body).Decode(&filters); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(filters["kldi"]) != 2 {
		t.Errorf("expected 2 kldi options, got %d", len(filters["kldi"]))
	}
}

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
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if result.TotalSelesai != 1 {
		t.Errorf("expected 1 selesai, got %d", result.TotalSelesai)
	}
}

func TestGetRealisasiWhenNotLoaded(t *testing.T) {
	svc := service.NewProcurementService(nil)
	h := NewHandler(svc)
	// realSvc not set

	req := httptest.NewRequest(http.MethodGet, "/api/v1/realisasi/summary", nil)
	w := httptest.NewRecorder()
	h.GetRealisasiSummary(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}
