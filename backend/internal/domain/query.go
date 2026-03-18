package domain

// ProcurementQuery holds filter, sort, and pagination parameters for a list query.
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

// PaginatedResult is the paginated response for a list query.
type PaginatedResult struct {
	Data       []Procurement `json:"data"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"pageSize"`
	TotalPages int           `json:"totalPages"`
}
