package domain

// JoinedRecord links a SIRUP RUP plan record with its SPSE execution result.
// If the package has not yet been tendered, Tender is nil.
type JoinedRecord struct {
	RUP    Procurement   `json:"rup"`
	Tender *TenderResult `json:"tender,omitempty"`
}

// RealisasiSummary provides aggregate budget utilisation metrics.
type RealisasiSummary struct {
	TotalPagu     float64 `json:"totalPagu"`
	TotalKontrak  float64 `json:"totalKontrak"`
	TotalSelesai  int     `json:"totalSelesai"`
	UtilisasiRate float64 `json:"utilisasiRate"` // totalKontrak / totalPagu * 100
	BelumTender   int     `json:"belumTender"`
}
