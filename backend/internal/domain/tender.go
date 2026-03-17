package domain

// TenderResult holds the execution outcome for one SPSE tender or pengadaan langsung package.
type TenderResult struct {
	ID             int64   `json:"id" parquet:"name=id, type=INT64"`
	NamaPaket      string  `json:"namaPaket" parquet:"name=nama_paket, type=BYTE_ARRAY, convertedtype=UTF8"`
	NilaiPagu      float64 `json:"nilaiPagu" parquet:"name=nilai_pagu, type=DOUBLE"`
	NilaiHPS       float64 `json:"nilaiHPS" parquet:"name=nilai_hps, type=DOUBLE"`
	NilaiKontrak   float64 `json:"nilaiKontrak" parquet:"name=nilai_kontrak, type=DOUBLE"`
	Tahap          string  `json:"tahap" parquet:"name=tahap, type=BYTE_ARRAY, convertedtype=UTF8"`
	SatuanKerja    string  `json:"satuanKerja" parquet:"name=satuan_kerja, type=BYTE_ARRAY, convertedtype=UTF8"`
	Jenis          string  `json:"jenis" parquet:"name=jenis, type=BYTE_ARRAY, convertedtype=UTF8"` // "lelang" | "pl"
	Pemenang       string  `json:"pemenang" parquet:"name=pemenang, type=BYTE_ARRAY, convertedtype=UTF8"`
	NilaiPenawaran float64 `json:"nilaiPenawaran" parquet:"name=nilai_penawaran, type=DOUBLE"`
	NPWP           string  `json:"npwp" parquet:"name=npwp, type=BYTE_ARRAY, convertedtype=UTF8"`
}
