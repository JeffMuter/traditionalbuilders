package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/emerald/traditionbuilders/internal/handlers"
	"github.com/emerald/traditionbuilders/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

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
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Recover runs inside LogRequests so panics are turned into a logged 500
	// that the request log can still observe.
	handler := handlers.LogRequests(handlers.Recover(mux))

	addr := ":8081"
	slog.Info("server starting", "url", "http://localhost"+addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
