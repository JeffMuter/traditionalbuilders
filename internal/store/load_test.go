package store

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestFindBuildersNear_ConcurrentSearches runs many concurrent proximity
// searches against a shared store to check for races and gross latency.
// Run with -race to catch data races.
func TestFindBuildersNear_ConcurrentSearches(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping load test in -short mode")
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	// SQLite in-memory is shared across the pool only via a single connection.
	db.SetMaxOpenConns(1)

	schema := `
	CREATE TABLE zip_codes (zip TEXT PRIMARY KEY, lat REAL NOT NULL, lng REAL NOT NULL, city TEXT NOT NULL, state TEXT NOT NULL);
	CREATE TABLE professionals (id INTEGER PRIMARY KEY, name TEXT, specialty TEXT, zip_code TEXT, latitude REAL, longitude REAL);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("schema: %v", err)
	}
	zips := []struct {
		zip      string
		lat, lng float64
		city, st string
	}{
		{"32801", 28.5383, -81.3792, "Orlando", "FL"},
		{"29401", 32.7765, -79.9311, "Charleston", "SC"},
		{"90210", 34.0901, -118.4065, "Beverly Hills", "CA"},
		{"37205", 36.0922, -86.8641, "Nashville", "TN"},
		{"19103", 39.9527, -75.1652, "Philadelphia", "PA"},
	}
	for _, z := range zips {
		if _, err := db.Exec(
			`INSERT INTO zip_codes (zip, lat, lng, city, state) VALUES (?, ?, ?, ?, ?)`,
			z.zip, z.lat, z.lng, z.city, z.st,
		); err != nil {
			t.Fatalf("seed zip: %v", err)
		}
	}
	// 50 professionals spread near each seeded zip.
	id := 1
	for _, z := range zips {
		for i := 0; i < 50; i++ {
			if _, err := db.Exec(
				`INSERT INTO professionals (id, name, specialty, zip_code, latitude, longitude) VALUES (?, ?, 'test', ?, ?, ?)`,
				id, "Pro", z.zip, z.lat+0.01*float64(i), z.lng+0.01*float64(i),
			); err != nil {
				t.Fatalf("seed pro: %v", err)
			}
			id++
		}
	}

	s := New(db)

	const workers = 100
	zipsToQuery := []string{"32801", "29401", "90210", "37205", "19103"}

	var wg sync.WaitGroup
	errs := make(chan error, workers)
	latencies := make([]time.Duration, workers)

	start := time.Now()
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			zip := zipsToQuery[w%len(zipsToQuery)]
			t0 := time.Now()
			results, err := s.FindBuildersNear(context.Background(), zip)
			latencies[w] = time.Since(t0)
			if err != nil {
				errs <- err
				return
			}
			if len(results) == 0 {
				errs <- &emptyResultsError{zip}
			}
		}(w)
	}
	wg.Wait()
	close(errs)
	total := time.Since(start)

	for err := range errs {
		t.Errorf("concurrent search error: %v", err)
	}

	// p95 latency check (single connection serializes, so allow generous bound).
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	p95 := sorted[int(float64(len(sorted))*0.95)]
	t.Logf("100 concurrent searches: total=%s p95=%s", total, p95)
	if p95 > 200*time.Millisecond {
		t.Errorf("p95 latency %s exceeds 200ms budget", p95)
	}
}

type emptyResultsError struct{ zip string }

func (e *emptyResultsError) Error() string {
	return "no results for zip " + e.zip
}
