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

func TestParseWinner(t *testing.T) {
	body := `[{"nama_peserta":"PT Maju Bersama","npwp":"12.345.678.9-000.000","nilai_penawaran":"4500000000","is_winner":"1"}]`
	winner, found := parseWinner([]byte(body))
	if !found {
		t.Fatal("expected winner found")
	}
	if winner.Pemenang != "PT Maju Bersama" {
		t.Errorf("expected 'PT Maju Bersama', got %q", winner.Pemenang)
	}
}
