package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/emerald/traditionbuilders/internal/logging"
	"github.com/emerald/traditionbuilders/internal/store"
	"github.com/emerald/traditionbuilders/templates"
)

// Profile handles GET /builders/{id} and renders a single builder's profile.
func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		renderStatus(w, r, http.StatusNotFound, templates.ProfileNotFound())
		return
	}

	pro, err := h.Store.GetProfessional(r.Context(), id)
	if errors.Is(err, store.ErrProfessionalNotFound) {
		renderStatus(w, r, http.StatusNotFound, templates.ProfileNotFound())
		return
	}
	if err != nil {
		serverError(w, r, "get professional", err)
		return
	}

	// A profile is still worth showing when the projects lookup fails; log the
	// failure and degrade to an empty list so the page stays useful.
	projects, err := h.Store.ListProjects(r.Context(), id)
	if err != nil {
		logging.FromContext(r.Context()).Error("list projects", "professional_id", id, "err", err)
		projects = nil
	}

	render(w, r, templates.Profile(pro, projects))
}
