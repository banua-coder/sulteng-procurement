package domain

// Summary holds aggregated analytics for a set of procurement records.
type Summary struct {
	TotalPagu  float64         `json:"totalPagu"`
	TotalPaket int             `json:"totalPaket"`
	JenisCount int             `json:"jenisCount"`
	KLDICount  int             `json:"kldiCount"`
	TopKLDI    string          `json:"topKldi"`
	ByJenis    []CategoryTotal `json:"byJenis"`
	ByKLDI     []CategoryTotal `json:"byKldi"`
	ByMetode   []CategoryTotal `json:"byMetode"`
	BySatker   []CategoryTotal `json:"bySatker"`
	TopItems   []Procurement   `json:"topItems"`
}

// CategoryTotal is a named bucket used in summary breakdowns.
type CategoryTotal struct {
	Name  string  `json:"name"`
	Total float64 `json:"total"`
	Count int     `json:"count"`
}
