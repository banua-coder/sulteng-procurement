package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/banua-coder/sulteng-procurement/backend/internal/domain"
	"github.com/banua-coder/sulteng-procurement/backend/internal/service"
)

// Handler holds the procurement service and exposes HTTP handlers.
type Handler struct {
	svc *service.ProcurementService
}

// NewHandler creates a handler backed by the given service.
func NewHandler(svc *service.ProcurementService) *Handler {
	return &Handler{svc: svc}
}

// GetSummary returns aggregated procurement statistics.
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.svc.GetSummary())
}

// GetProcurements returns a filtered, sorted, paginated list of procurement records.
func (h *Handler) GetProcurements(w http.ResponseWriter, r *http.Request) {
	q := domain.ProcurementQuery{
		Page:           intParam(r, "page", 1),
		PageSize:       intParam(r, "pageSize", 25),
		Search:         r.URL.Query().Get("search"),
		KLDI:           r.URL.Query().Get("kldi"),
		JenisPengadaan: r.URL.Query().Get("jenisPengadaan"),
		Metode:         r.URL.Query().Get("metode"),
		SortBy:         r.URL.Query().Get("sortBy"),
		SortDir:        r.URL.Query().Get("sortDir"),
	}
	writeJSON(w, h.svc.Query(q))
}

// GetFilters returns the distinct values available for each filter field.
func (h *Handler) GetFilters(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.svc.GetFilters())
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func intParam(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
