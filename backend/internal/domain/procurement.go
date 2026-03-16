package domain

type Procurement struct {
	ID               int64   `json:"id" parquet:"name=id, type=INT64"`
	Paket            string  `json:"paket" parquet:"name=paket, type=BYTE_ARRAY, convertedtype=UTF8"`
	Pagu             float64 `json:"pagu" parquet:"name=pagu, type=DOUBLE"`
	JenisPengadaan   string  `json:"jenisPengadaan" parquet:"name=jenis_pengadaan, type=BYTE_ARRAY, convertedtype=UTF8"`
	Metode           string  `json:"metode" parquet:"name=metode, type=BYTE_ARRAY, convertedtype=UTF8"`
	Pemilihan        string  `json:"pemilihan" parquet:"name=pemilihan, type=BYTE_ARRAY, convertedtype=UTF8"`
	SatuanKerja      string  `json:"satuanKerja" parquet:"name=satuan_kerja, type=BYTE_ARRAY, convertedtype=UTF8"`
	KLDI             string  `json:"kldi" parquet:"name=kldi, type=BYTE_ARRAY, convertedtype=UTF8"`
	Lokasi           string  `json:"lokasi" parquet:"name=lokasi, type=BYTE_ARRAY, convertedtype=UTF8"`
	SumberDana       string  `json:"sumberDana" parquet:"name=sumber_dana, type=BYTE_ARRAY, convertedtype=UTF8"`
	IsPDN            bool    `json:"isPDN" parquet:"name=is_pdn, type=BOOLEAN"`
	IsUMK            bool    `json:"isUMK" parquet:"name=is_umk, type=BOOLEAN"`
	IdBulan          int     `json:"idBulan" parquet:"name=id_bulan, type=INT32"`
	IdKldi           string  `json:"idKldi" parquet:"name=id_kldi, type=BYTE_ARRAY, convertedtype=UTF8"`
	IdReferensi      int     `json:"idReferensi" parquet:"name=id_referensi, type=INT32"`
	IdSatker         int     `json:"idSatker" parquet:"name=id_satker, type=INT32"`
	IdMetode         int     `json:"idMetode" parquet:"name=id_metode, type=INT32"`
	IdJenisPengadaan int     `json:"idJenisPengadaan" parquet:"name=id_jenis_pengadaan, type=INT32"`
}

type Summary struct {
	TotalPagu  float64         `json:"totalPagu"`
	TotalPaket int             `json:"totalPaket"`
	JenisCount int             `json:"jenisCount"`
	KLDICount  int             `json:"kldiCount"`
	TopKLDI    string          `json:"topKldi"`
	ByJenis    []CategoryTotal `json:"byJenis"`
	ByKLDI     []CategoryTotal `json:"byKldi"`
	ByMetode   []CategoryTotal `json:"byMetode"`
}

type CategoryTotal struct {
	Name  string  `json:"name"`
	Total float64 `json:"total"`
	Count int     `json:"count"`
}

type ProcurementQuery struct {
	Page           int    `json:"page"`
	PageSize       int    `json:"pageSize"`
	Search         string `json:"search"`
	KLDI           string `json:"kldi"`
	JenisPengadaan string `json:"jenisPengadaan"`
	Metode         string `json:"metode"`
	SortBy         string `json:"sortBy"`
	SortDir        string `json:"sortDir"`
}

type PaginatedResult struct {
	Data       []Procurement `json:"data"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"pageSize"`
	TotalPages int           `json:"totalPages"`
}
