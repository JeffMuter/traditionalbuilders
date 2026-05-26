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

func Landing(w http.ResponseWriter, r *http.Request) {
	templates.Landing().Render(r.Context(), w)
}

func Gallery(w http.ResponseWriter, r *http.Request) {
	templates.Gallery().Render(r.Context(), w)
}
