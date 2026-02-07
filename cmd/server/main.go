package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"github.com/emerald/traditionbuilders/internal/handlers"
)

func main() {
	// Initialize database
	db, err := sql.Open("sqlite3", "./traditionbuilders.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Database connection established")

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Landing)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start server
	addr := ":8080"
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
