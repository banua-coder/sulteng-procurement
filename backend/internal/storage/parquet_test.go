package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ryanaidilp/sulteng-procurement/backend/internal/domain"
)

func TestWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	store := NewParquetStore(dir)

	records := []domain.Procurement{
		{ID: 1, Paket: "Package A", Pagu: 1000000, JenisPengadaan: "Barang", KLDI: "Kota Palu", Metode: "Tender"},
		{ID: 2, Paket: "Package B", Pagu: 2000000, JenisPengadaan: "Jasa Lainnya", KLDI: "Kab. Poso", Metode: "E-Purchasing"},
	}

	err := store.Write(2026, records)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	path := filepath.Join(dir, "procurement_2026.parquet")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("parquet file was not created")
	}

	loaded, err := store.Read(2026)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if len(loaded) != 2 {
		t.Errorf("expected 2 records, got %d", len(loaded))
	}
	if loaded[0].Paket != "Package A" {
		t.Errorf("expected 'Package A', got %s", loaded[0].Paket)
	}
	if loaded[1].Pagu != 2000000 {
		t.Errorf("expected pagu=2000000, got %f", loaded[1].Pagu)
	}
}
