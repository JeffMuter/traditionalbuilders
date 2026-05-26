package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"github.com/emerald/traditionbuilders/internal/handlers"
	"github.com/emerald/traditionbuilders/internal/store"
)

func main() {
	// Initialize database
	db, err := sql.Open("sqlite3", "./traditionbuilders.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Println("Database connection established")

	// Store encapsulates all DB queries; handler receives the store, not the raw DB.
	s := store.New(db)
	h := &handlers.Handler{Store: s}

	// Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Landing)
	mux.HandleFunc("/gallery", handlers.Gallery)
	mux.HandleFunc("/builders", h.BuildersPage)
	mux.HandleFunc("/api/builders", h.SearchBuilders)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	addr := ":8081"
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
