package storage

import (
	"fmt"
	"path/filepath"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/writer"

	"github.com/banua-coder/sulteng-procurement/backend/internal/domain"
)

// ParquetStore reads and writes procurement records as Parquet files.
type ParquetStore struct {
	dir string
}

// NewParquetStore creates a store that persists files under dir.
func NewParquetStore(dir string) *ParquetStore {
	return &ParquetStore{dir: dir}
}

func (s *ParquetStore) path(year int) string {
	return filepath.Join(s.dir, fmt.Sprintf("procurement_%d.parquet", year))
}

// Write serialises records to a Parquet file for the given year.
func (s *ParquetStore) Write(year int, records []domain.Procurement) error {
	fw, err := local.NewLocalFileWriter(s.path(year))
	if err != nil {
		return fmt.Errorf("create file writer: %w", err)
	}

	pw, err := writer.NewParquetWriter(fw, new(domain.Procurement), 4)
	if err != nil {
		return fmt.Errorf("create parquet writer: %w", err)
	}
	pw.RowGroupSize = 128 * 1024 * 1024
	pw.CompressionType = 1 // SNAPPY

	for _, r := range records {
		if err := pw.Write(r); err != nil {
			return fmt.Errorf("write record: %w", err)
		}
	}

	if err := pw.WriteStop(); err != nil {
		return fmt.Errorf("finalize parquet: %w", err)
	}
	return fw.Close()
}

// Read deserialises all records from the Parquet file for the given year.
func (s *ParquetStore) Read(year int) ([]domain.Procurement, error) {
	fr, err := local.NewLocalFileReader(s.path(year))
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = fr.Close() }()

	pr, err := reader.NewParquetReader(fr, new(domain.Procurement), 4)
	if err != nil {
		return nil, fmt.Errorf("create reader: %w", err)
	}
	defer pr.ReadStop()

	num := int(pr.GetNumRows())
	records := make([]domain.Procurement, num)
	if err := pr.Read(&records); err != nil {
		return nil, fmt.Errorf("read records: %w", err)
	}

	return records, nil
}

// Exists reports whether a Parquet file for the given year exists.
func (s *ParquetStore) Exists(year int) bool {
	_, err := local.NewLocalFileReader(s.path(year))
	return err == nil
}
