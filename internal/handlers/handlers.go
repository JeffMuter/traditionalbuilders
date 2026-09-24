package handlers

import (
	"net/http"

	"github.com/emerald/traditionbuilders/internal/store"
	"github.com/emerald/traditionbuilders/templates"
)

// Handler holds the dependencies shared by all DB-backed HTTP handlers.
// Plain page handlers (Landing, Gallery) remain package-level functions.
type Handler struct {
	Store *store.Store
}

// Landing serves the home page. Any other unmatched path falls through to the
// "/" pattern, so it is answered with a styled 404 instead of the home page.
func Landing(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		renderStatus(w, r, http.StatusNotFound, templates.ErrorPage(
			http.StatusNotFound,
			"Page not found",
			"The page you're looking for doesn't exist or has moved.",
		))
		return
	}
	render(w, r, templates.Landing())
}

func Gallery(w http.ResponseWriter, r *http.Request) {
	render(w, r, templates.Gallery())
}
