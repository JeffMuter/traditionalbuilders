// seed-zips downloads the GeoNames US postal code dataset and bulk-imports
// it into the zip_codes table, replacing any previous data.
//
// Usage:
//
//	go run cmd/seed-zips/main.go
//	go run cmd/seed-zips/main.go path/to/traditionbuilders.db
package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	geonamesURL = "https://download.geonames.org/export/zip/US.zip"
	defaultDB   = "traditionbuilders.db"
)

func main() {
	dbPath := defaultDB
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	log.Printf("Downloading GeoNames US postal code data (~3 MB)…")
	data, err := downloadFile(geonamesURL)
	if err != nil {
		log.Fatalf("download: %v", err)
	}

	log.Printf("Parsing archive…")
	rows, err := parseGeonames(data)
	if err != nil {
		log.Fatalf("parse: %v", err)
	}
	log.Printf("Parsed %d zip codes", len(rows))

	log.Printf("Importing into %s…", dbPath)
	if err := bulkImport(dbPath, rows); err != nil {
		log.Fatalf("import: %v", err)
	}

	log.Printf("Done — %d zip codes loaded.", len(rows))
}

type zipRow struct {
	Zip   string
	Lat   float64
	Lng   float64
	City  string
	State string
}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

// parseGeonames reads US.txt from the in-memory zip archive.
// GeoNames tab-separated columns (1-indexed):
//
//	1  country_code
//	2  postal_code
//	3  place_name
//	4  admin_name1  (full state name)
//	5  admin_code1  (state abbreviation, e.g. FL)
//	6–9 county/district fields
//	10 latitude
//	11 longitude
//	12 accuracy
func parseGeonames(data []byte) ([]zipRow, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	var txtFile *zip.File
	for _, f := range zr.File {
		if f.Name == "US.txt" {
			txtFile = f
			break
		}
	}
	if txtFile == nil {
		return nil, fmt.Errorf("US.txt not found in archive")
	}

	rc, err := txtFile.Open()
	if err != nil {
		return nil, fmt.Errorf("open US.txt: %w", err)
	}
	defer rc.Close()

	var rows []zipRow
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) < 11 {
			continue
		}
		lat, err := strconv.ParseFloat(parts[9], 64)
		if err != nil {
			continue
		}
		lng, err := strconv.ParseFloat(parts[10], 64)
		if err != nil {
			continue
		}
		rows = append(rows, zipRow{
			Zip:   parts[1],
			Lat:   lat,
			Lng:   lng,
			City:  parts[2],
			State: parts[4],
		})
	}
	return rows, scanner.Err()
}

// bulkImport replaces all zip_codes rows inside a single transaction.
func bulkImport(dbPath string, rows []zipRow) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM zip_codes"); err != nil {
		return fmt.Errorf("clear table: %w", err)
	}

	stmt, err := tx.Prepare(
		`INSERT OR IGNORE INTO zip_codes (zip, lat, lng, city, state) VALUES (?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range rows {
		if _, err := stmt.Exec(r.Zip, r.Lat, r.Lng, r.City, r.State); err != nil {
			return fmt.Errorf("insert %s: %w", r.Zip, err)
		}
	}

	return tx.Commit()
}
