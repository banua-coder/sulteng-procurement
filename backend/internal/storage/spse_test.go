package storage

import (
	"os"
	"testing"

	"github.com/banua-coder/sulteng-procurement/backend/internal/domain"
)

func TestWriteAndReadTender(t *testing.T) {
	dir := t.TempDir()
	store := NewSpseStore(dir)

	input := []domain.TenderResult{
		{ID: 1, NamaPaket: "Paket A", NilaiKontrak: 500000000, Pemenang: "PT Contoh", Jenis: "lelang"},
		{ID: 2, NamaPaket: "Paket B", NilaiKontrak: 200000000, Pemenang: "CV Maju", Jenis: "pl"},
	}

	if err := store.Write(2026, input); err != nil {
		t.Fatalf("Write: %v", err)
	}
	got, err := store.Read(2026)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 records, got %d", len(got))
	}
	if got[0].Pemenang != "PT Contoh" {
		t.Errorf("expected Pemenang='PT Contoh', got %q", got[0].Pemenang)
	}

	path := dir + "/spse_2026.parquet"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("parquet file was not created")
	}
}
