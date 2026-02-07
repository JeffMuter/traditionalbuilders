package handlers

import (
	"net/http"

	"github.com/emerald/traditionbuilders/templates"
)

func Landing(w http.ResponseWriter, r *http.Request) {
	templates.Landing().Render(r.Context(), w)
}

func Gallery(w http.ResponseWriter, r *http.Request) {
	templates.Gallery().Render(r.Context(), w)
}
