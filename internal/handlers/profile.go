package handlers

import (
	"net/http"
	"strconv"

	"github.com/emerald/traditionbuilders/internal/store"
	"github.com/emerald/traditionbuilders/templates"
)

// Profile handles GET /builders/{id} and renders a single builder's profile.
func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		w.WriteHeader(http.StatusNotFound)
		templates.ProfileNotFound().Render(r.Context(), w)
		return
	}

	pro, err := h.Store.GetProfessional(r.Context(), id)
	if err == store.ErrProfessionalNotFound {
		w.WriteHeader(http.StatusNotFound)
		templates.ProfileNotFound().Render(r.Context(), w)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// A profile is still worth showing when the projects lookup fails; log-free
	// degradation to an empty list keeps the page useful.
	projects, err := h.Store.ListProjects(r.Context(), id)
	if err != nil {
		projects = nil
	}

	templates.Profile(pro, projects).Render(r.Context(), w)
}
