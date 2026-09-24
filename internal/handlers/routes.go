package handlers

import (
	"database/sql"
	"net/http"

	"github.com/emerald/traditionbuilders/internal/logging"
)

// Routes builds the full application handler: every route, the static file
// server, and the request-scoped middleware. It is the single source of truth
// shared by cmd/server, the render-contract tests, and the browser E2E suite so
// a route cannot exist in production but not in tests (or vice versa).
//
// staticDir is the on-disk directory served under /static/. Tests pass the
// repo-relative path because their working directory differs from production.
func (h *Handler) Routes(db *sql.DB, staticDir string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", Landing)
	mux.HandleFunc("/gallery", Gallery)
	mux.HandleFunc("/builders", h.BuildersPage)
	mux.HandleFunc("/builders/{id}", h.Profile)
	mux.HandleFunc("/api/builders", h.SearchBuilders)
	mux.HandleFunc("/healthz", healthz(db))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// RequestLogger runs outermost so the request ID and scoped logger exist
	// before Recover, letting panics be logged with the same correlation ID.
	return RequestLogger(Recover(mux))
}

// healthz reports readiness by pinging the database. A ping failure returns
// 503 so a load balancer can drain the instance.
func healthz(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			logging.FromContext(r.Context()).Error("health check failed", "err", err)
			http.Error(w, "unhealthy", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	}
}
