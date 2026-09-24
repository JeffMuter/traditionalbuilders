// seed-zips loads the embedded GeoNames US postal-code dataset into the
// zip_codes table.
//
// The dataset is vendored in internal/zipdata (see
// db/scripts/build-zip-dataset.sh), so this command is fully offline and
// reproducible. It is idempotent: if the embedded dataset's SHA-256 already
// matches the data_seeds row, it does nothing.
//
// Usage:
//
//	go run ./cmd/seed-zips
//	go run ./cmd/seed-zips path/to/traditionbuilders.db
package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/emerald/traditionbuilders/internal/zipdata"
)

const defaultDB = "traditionbuilders.db"

func main() {
	dbPath := defaultDB
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("open %s: %v", dbPath, err)
	}
	defer db.Close()

	applied, err := zipdata.EnsureLoaded(db)
	if err != nil {
		log.Fatalf("seed zip codes: %v", err)
	}
	if !applied {
		log.Printf("zip codes already up to date for %s", dbPath)
		return
	}
	log.Printf("Loaded the full zip code dataset into %s", dbPath)
}
