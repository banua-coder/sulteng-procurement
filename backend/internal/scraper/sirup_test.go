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
