package handlers

import (
	"errors"
	"net/http"

	"github.com/emerald/traditionbuilders/internal/store"
	"github.com/emerald/traditionbuilders/templates"
)

// BuildersPage handles GET /builders?zip=XXXXX and renders the HTML results page.
func (h *Handler) BuildersPage(w http.ResponseWriter, r *http.Request) {
	zip := r.URL.Query().Get("zip")

	if zip == "" || !zipPattern.MatchString(zip) {
		render(w, r, templates.Builders(zip, nil,
			"Please enter a valid 5-digit zip code.",
			"Please return to the search page and enter a valid 5-digit US zip code."))
		return
	}

	results, err := h.Store.FindBuildersNear(r.Context(), baseZip(zip))
	if errors.Is(err, store.ErrZipNotFound) {
		render(w, r, templates.Builders(zip, nil,
			"We don't recognize zip code "+zip+".",
			"Try a nearby zip code, or ask your builder to register."))
		return
	}
	if err != nil {
		serverError(w, r, "find builders near zip", err)
		return
	}

	render(w, r, templates.Builders(zip, results, "", ""))
}
