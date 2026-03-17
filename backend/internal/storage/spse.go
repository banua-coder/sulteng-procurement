package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xitongsys/parquet-go-source/local"
	goparquet "github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/writer"

	"github.com/banua-coder/sulteng-procurement/backend/internal/domain"
)

// SpseStore persists SPSE tender results as parquet files.
type SpseStore struct {
	dir string
}

// NewSpseStore creates a store that persists files under dir.
func NewSpseStore(dir string) *SpseStore {
	return &SpseStore{dir: dir}
}

func (s *SpseStore) path(year int) string {
	return filepath.Join(s.dir, fmt.Sprintf("spse_%d.parquet", year))
}

// Exists reports whether a parquet file for the given year exists.
func (s *SpseStore) Exists(year int) bool {
	_, err := os.Stat(s.path(year))
	return err == nil
}

// Write serialises records to a parquet file for the given year.
func (s *SpseStore) Write(year int, records []domain.TenderResult) error {
	fw, err := local.NewLocalFileWriter(s.path(year))
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	pw, err := writer.NewParquetWriter(fw, new(domain.TenderResult), 4)
	if err != nil {
		return fmt.Errorf("new writer: %w", err)
	}
	pw.CompressionType = goparquet.CompressionCodec_SNAPPY

	for _, r := range records {
		if err := pw.Write(r); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}

	if err := pw.WriteStop(); err != nil {
		return fmt.Errorf("write stop: %w", err)
	}
	return fw.Close()
}

// Read deserialises all records from the parquet file for the given year.
func (s *SpseStore) Read(year int) ([]domain.TenderResult, error) {
	fr, err := local.NewLocalFileReader(s.path(year))
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer fr.Close()

	pr, err := reader.NewParquetReader(fr, new(domain.TenderResult), 4)
	if err != nil {
		return nil, fmt.Errorf("new reader: %w", err)
	}
	defer pr.ReadStop()

	n := int(pr.GetNumRows())
	records := make([]domain.TenderResult, n)
	if err := pr.Read(&records); err != nil {
		return nil, fmt.Errorf("read rows: %w", err)
	}
	return records, nil
}
