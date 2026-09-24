package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/emerald/traditionbuilders/internal/handlers"
	"github.com/emerald/traditionbuilders/internal/logging"
	"github.com/emerald/traditionbuilders/internal/store"
)

// logLevelFromEnv maps LOG_LEVEL (debug|info|warn|error) to an slog.Level.
// Unset or unrecognised values fall back to INFO.
func logLevelFromEnv() slog.Level {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL"))) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func main() {
	level := logLevelFromEnv()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))
	slog.Info("logging configured", "level", level.String())

	// Initialize database
	db, err := sql.Open("sqlite3", "./traditionbuilders.db")
	if err != nil {
		slog.Error("open database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("ping database", "err", err)
		os.Exit(1)
	}
	slog.Info("database connection established")

	// Store encapsulates all DB queries; handler receives the store, not the raw DB.
	s := store.New(db)
	h := &handlers.Handler{Store: s}

	// Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Landing)
	mux.HandleFunc("/gallery", handlers.Gallery)
	mux.HandleFunc("/builders", h.BuildersPage)
	mux.HandleFunc("/builders/{id}", h.Profile)
	mux.HandleFunc("/api/builders", h.SearchBuilders)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			logging.FromContext(r.Context()).Error("health check failed", "err", err)
			http.Error(w, "unhealthy", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// RequestLogger runs outermost so the request ID and scoped logger exist
	// before Recover, letting panics be logged with the same correlation ID.
	handler := handlers.RequestLogger(handlers.Recover(mux))

	addr := ":8081"
	if v := strings.TrimSpace(os.Getenv("ADDR")); v != "" {
		addr = v
	}
	slog.Info("server starting", "url", "http://localhost"+addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
