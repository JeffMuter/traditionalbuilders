package main

import (
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/emerald/traditionbuilders/internal/handlers"
	"github.com/emerald/traditionbuilders/internal/store"
	"github.com/emerald/traditionbuilders/internal/zipdata"
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

	// -seed-only applies the embedded zip dataset and exits. Deploy runs this
	// after migrations so prod gets the full dataset without the web server
	// depending on network access or a manual step.
	seedOnly := flag.Bool("seed-only", false, "apply the embedded zip code dataset and exit")
	flag.Parse()

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

	if *seedOnly {
		if _, err := zipdata.EnsureLoaded(db); err != nil {
			slog.Error("seed zip codes", "err", err)
			os.Exit(1)
		}
		slog.Info("zip code dataset ready")
		return
	}

	// Best-effort seed on normal startup: keeps a fresh dev/prod DB complete
	// without a separate step. A failure (e.g. migrations not run yet) is
	// logged, not fatal, so the server still starts and reports unhealthy.
	if applied, err := zipdata.EnsureLoaded(db); err != nil {
		slog.Warn("zip code seed skipped", "err", err)
	} else if applied {
		slog.Info("zip code dataset loaded")
	} else {
		slog.Info("zip code dataset already current")
	}

	// Store encapsulates all DB queries; handler receives the store, not the raw DB.
	s := store.New(db)
	h := &handlers.Handler{Store: s}

	// Routes live in the handlers package so production, the render-contract
	// tests, and the browser E2E suite all exercise the same mux.
	handler := h.Routes(db, "static")

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
