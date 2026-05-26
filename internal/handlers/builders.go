package handlers

import (
	"net/http"

	"github.com/emerald/traditionbuilders/internal/store"
	"github.com/emerald/traditionbuilders/templates"
)

// BuildersPage handles GET /builders?zip=XXXXX and renders the HTML results page.
func (h *Handler) BuildersPage(w http.ResponseWriter, r *http.Request) {
	zip := r.URL.Query().Get("zip")

	if zip == "" || !zipPattern.MatchString(zip) {
		templates.Builders(zip, nil, "Please enter a valid 5-digit zip code.").Render(r.Context(), w)
		return
	}

	results, err := h.Store.FindBuildersNear(r.Context(), zip)
	if err == store.ErrZipNotFound {
		templates.Builders(zip, nil, "We don't recognize zip code "+zip+". Try a nearby zip code, or ask your builder to register.").Render(r.Context(), w)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	templates.Builders(zip, results, "").Render(r.Context(), w)
}
