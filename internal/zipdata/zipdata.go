// Package zipdata embeds the vendored GeoNames US postal-code dataset and
// loads it into the zip_codes table.
//
// The dataset is generated offline by db/scripts/build-zip-dataset.sh. It is
// committed as a gzipped TSV and embedded into the server binary, so the seed
// path has no network dependency and is reproducible from a checked-out tree.
package zipdata

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// DatasetName identifies the seed in the data_seeds table.
const DatasetName = "zip_codes"

//go:embed zip_codes.tsv.gz
var rawDataset []byte

//go:embed dataset.json
var rawMeta []byte

// Manifest describes the vendored dataset. It is parsed from dataset.json
// (written by the build script) and travels with the binary.
type Manifest struct {
	Name        string   `json:"name"`
	Source      string   `json:"source"`
	SourceURL   string   `json:"source_url"`
	License     string   `json:"license"`
	Attribution string   `json:"attribution"`
	FetchedAt   string   `json:"fetched_at"`
	RowCount    int      `json:"row_count"`
	Columns     []string `json:"columns"`
	SHA256      string   `json:"sha256"`
}

// Row is one postal-code record.
type Row struct {
	Zip   string
	Lat   float64
	Lng   float64
	City  string
	State string
}

// Sanity bounds for continental US + AK/HI; corrupt rows are rejected.
const (
	minLat = 15.0
	maxLat = 72.0
	minLng = -180.0
	maxLng = -60.0
)

// Metadata returns the embedded dataset manifest.
func Metadata() (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(rawMeta, &m); err != nil {
		return Manifest{}, fmt.Errorf("parse embedded dataset.json: %w", err)
	}
	return m, nil
}

// Each streams every well-formed row to fn. It never materializes the full
// dataset in memory and stops at the first error returned by fn.
func Each(fn func(Row) error) error {
	gr, err := gzip.NewReader(bytes.NewReader(rawDataset))
	if err != nil {
		return fmt.Errorf("open embedded dataset: %w", err)
	}
	defer gr.Close()

	scanner := bufio.NewScanner(gr)
	// 64 KB is plenty for a single TSV line; widen defensively.
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		row, ok := parseLine(scanner.Text())
		if !ok {
			continue
		}
		if err := fn(row); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read embedded dataset: %w", err)
	}
	return nil
}

// Count returns the number of well-formed rows after integrity filtering.
func Count() (int, error) {
	n := 0
	if err := Each(func(Row) error { n++; return nil }); err != nil {
		return 0, err
	}
	return n, nil
}

// parseLine parses one "zip\tlat\tlng\tcity\tstate" line. It returns ok=false
// for malformed or out-of-range rows so the caller can skip them.
func parseLine(s string) (Row, bool) {
	parts := strings.Split(s, "\t")
	if len(parts) < 5 {
		return Row{}, false
	}
	zip := parts[0]
	if len(zip) != 5 || !allDigits(zip) {
		return Row{}, false
	}
	lat, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return Row{}, false
	}
	lng, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return Row{}, false
	}
	if lat < minLat || lat > maxLat || lng < minLng || lng > maxLng {
		return Row{}, false
	}
	city, state := parts[3], parts[4]
	// Military FPO/APO rows have no state; skip them so the table stays
	// consistent with the vendored dataset's normalization.
	if city == "" || state == "" {
		return Row{}, false
	}
	return Row{Zip: zip, Lat: lat, Lng: lng, City: city, State: state}, true
}

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Loaded reports whether the dataset version currently embedded in the binary
// has already been applied to db. It returns (false, nil) when the bookkeeping
// table or database is not ready yet.
func Loaded(db *sql.DB, sha string) (bool, error) {
	var stored string
	err := db.QueryRow(
		`SELECT sha256 FROM data_seeds WHERE name = ?`, DatasetName,
	).Scan(&stored)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		// Table missing (migrations not run) or other read error: treat as not
		// loaded so the caller can decide; report the error for logging.
		return false, err
	}
	return stored == sha, nil
}

// EnsureLoaded applies the embedded dataset to db unless the exact version is
// already recorded in data_seeds. It is idempotent and safe to call on every
// startup.
//
// When the zip_codes table does not exist (migrations not yet run), it returns
// an error the caller is expected to log and continue past — the web server
// must not fail to start because a seed could not run.
func EnsureLoaded(db *sql.DB) (applied bool, err error) {
	meta, err := Metadata()
	if err != nil {
		return false, err
	}

	loaded, err := Loaded(db, meta.SHA256)
	if err == nil && loaded {
		return false, nil
	}
	// If Loaded errored because data_seeds is missing, fall through and let
	// the transaction surface a clear error.

	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM zip_codes`); err != nil {
		return false, fmt.Errorf("clear zip_codes: %w", err)
	}

	stmt, err := tx.Prepare(
		`INSERT OR IGNORE INTO zip_codes (zip, lat, lng, city, state) VALUES (?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	inserted := 0
	if err := Each(func(r Row) error {
		if _, err := stmt.Exec(r.Zip, r.Lat, r.Lng, r.City, r.State); err != nil {
			return fmt.Errorf("insert %s: %w", r.Zip, err)
		}
		inserted++
		return nil
	}); err != nil {
		return false, err
	}

	if _, err := tx.Exec(`
		INSERT INTO data_seeds (name, sha256, source, row_count, loaded_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(name) DO UPDATE SET
			sha256 = excluded.sha256,
			source = excluded.source,
			row_count = excluded.row_count,
			loaded_at = excluded.loaded_at
	`, DatasetName, meta.SHA256, meta.SourceURL, inserted); err != nil {
		return false, fmt.Errorf("record data_seeds: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
