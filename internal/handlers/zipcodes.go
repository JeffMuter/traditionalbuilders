package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/emerald/traditionbuilders/internal/logging"
	"github.com/emerald/traditionbuilders/internal/store"
)

// zipPattern accepts a 5-digit ZIP or a ZIP+4 (e.g. "32801" or "32801-1234").
// Search always uses the 5-digit base, so callers should normalize with
// baseZip before querying the store.
var zipPattern = regexp.MustCompile(`^\d{5}(?:-\d{4})?$`)

// writeError writes an error response with the appropriate HTTP status code.
// It is reached only after the zip-format check has passed, so the possible
// inputs are ErrZipNotFound (406), DB/context errors (500), or unknown errors
// (500). The status mapping lives in store.ToHTTPStatus.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status := store.ToHTTPStatus(err)
	msg := "internal error"
	if store.IsZipNotFound(err) {
		msg = "zip not found"
	} else {
		// Unexpected store failures are logged server-side; the client only
		// sees the generic message.
		logging.FromContext(r.Context()).Error("search builders failed", "status", status, "err", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if encErr := json.NewEncoder(w).Encode(map[string]string{"error": msg}); encErr != nil {
		logging.FromContext(r.Context()).Error("encode error response", "err", encErr)
	}
}

// baseZip strips any ZIP+4 extension, returning the 5-digit base zip.
// Non-matching input is returned unchanged.
func baseZip(zip string) string {
	if len(zip) > 5 {
		return zip[:5]
	}
	return zip
}

// SearchBuilders handles GET /api/builders?zip=XXXXX and returns JSON.
func (h *Handler) SearchBuilders(w http.ResponseWriter, r *http.Request) {
	zip := r.URL.Query().Get("zip")
	if !zipPattern.MatchString(zip) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid zip code"})
		return
	}

	results, err := h.Store.FindBuildersNear(r.Context(), baseZip(zip))
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results) // always a non-nil slice — never "null"
}
