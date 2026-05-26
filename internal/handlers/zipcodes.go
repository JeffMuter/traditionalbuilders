package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/emerald/traditionbuilders/internal/store"
)

var zipPattern = regexp.MustCompile(`^\d{5}$`)

// SearchBuilders handles GET /api/builders?zip=XXXXX and returns JSON.
func (h *Handler) SearchBuilders(w http.ResponseWriter, r *http.Request) {
	zip := r.URL.Query().Get("zip")
	if !zipPattern.MatchString(zip) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid zip code"})
		return
	}

	results, err := h.Store.FindBuildersNear(r.Context(), zip)
	if err != nil && err != store.ErrZipNotFound {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results) // always a non-nil slice — never "null"
}
