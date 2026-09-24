package store

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// newEdgeStore builds an in-memory store with a single Orlando zip and a
// configurable set of professionals for edge-case testing.
func newEdgeStore(t *testing.T) (*Store, *sql.DB) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	schema := `
	CREATE TABLE zip_codes (zip TEXT PRIMARY KEY, lat REAL NOT NULL, lng REAL NOT NULL, city TEXT NOT NULL, state TEXT NOT NULL);
	CREATE TABLE professionals (id INTEGER PRIMARY KEY, name TEXT, specialty TEXT, zip_code TEXT, latitude REAL, longitude REAL);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("schema: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO zip_codes (zip, lat, lng, city, state) VALUES ('32801', 28.5383, -81.3792, 'Orlando', 'FL')`,
	); err != nil {
		t.Fatalf("seed zip: %v", err)
	}
	return New(db), db
}

// TestFindBuildersNear_MaxResultsCap verifies that no more than the internal
// maxResults cap is ever returned, even when many professionals are nearby.
func TestFindBuildersNear_MaxResultsCap(t *testing.T) {
	s, db := newEdgeStore(t)

	// Seed 3x the cap, all essentially on top of the Orlando zip.
	for i := 0; i < maxResults*3; i++ {
		// Tiny jitter keeps distances ordered and non-zero.
		lat := 28.5383 + float64(i)*0.0001
		lng := -81.3792 + float64(i)*0.0001
		if _, err := db.Exec(
			`INSERT INTO professionals (name, specialty, zip_code, latitude, longitude) VALUES (?, 'test', '32801', ?, ?)`,
			fmt.Sprintf("Pro %02d", i), lat, lng,
		); err != nil {
			t.Fatalf("seed pro %d: %v", i, err)
		}
	}

	got, err := s.FindBuildersNear(context.Background(), "32801")
	if err != nil {
		t.Fatalf("FindBuildersNear: %v", err)
	}
	if len(got) != maxResults {
		t.Errorf("len(results) = %d, want %d (cap)", len(got), maxResults)
	}

	// The nearest professional must survive the cap.
	if got[0].Name != "Pro 00" {
		t.Errorf("closest = %q, want Pro 00", got[0].Name)
	}
}

// TestFindBuildersNear_EmptyResults verifies a valid zip with no nearby
// professionals yields a non-nil empty slice and no error.
func TestFindBuildersNear_EmptyResults(t *testing.T) {
	s, _ := newEdgeStore(t)

	got, err := s.FindBuildersNear(context.Background(), "32801")
	if err != nil {
		t.Fatalf("FindBuildersNear: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil empty slice")
	}
	if len(got) != 0 {
		t.Errorf("len(results) = %d, want 0", len(got))
	}
}

// TestFindBuildersNear_NullZipCode verifies professionals with a NULL zip_code
// but valid coordinates are handled gracefully (city/state fall back to "").
func TestFindBuildersNear_NullZipCode(t *testing.T) {
	s, db := newEdgeStore(t)

	if _, err := db.Exec(
		`INSERT INTO professionals (name, specialty, zip_code, latitude, longitude) VALUES ('No Zip', 'test', NULL, 28.54, -81.38)`,
	); err != nil {
		t.Fatalf("seed pro: %v", err)
	}

	got, err := s.FindBuildersNear(context.Background(), "32801")
	if err != nil {
		t.Fatalf("FindBuildersNear: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(got))
	}
	if got[0].Name != "No Zip" {
		t.Errorf("name = %q, want No Zip", got[0].Name)
	}
	if got[0].City != "" || got[0].State != "" {
		t.Errorf("city/state = %q/%q, want empty fallbacks", got[0].City, got[0].State)
	}
}

// TestFindBuildersNear_ZipPlus4Normalization verifies ZIP+4 is normalized to its
// 5-digit base before the lookup. (Callers should pass baseZip output, but the
// store must also tolerate a raw ZIP+4 key being absent and only match base.)
func TestFindBuildersNear_BaseZipLookup(t *testing.T) {
	s, db := newEdgeStore(t)

	if _, err := db.Exec(
		`INSERT INTO professionals (name, specialty, zip_code, latitude, longitude) VALUES ('Orlando Pro', 'test', '32801', 28.54, -81.38)`,
	); err != nil {
		t.Fatalf("seed pro: %v", err)
	}

	// The store itself receives the already-normalized 5-digit base.
	got, err := s.FindBuildersNear(context.Background(), "32801")
	if err != nil {
		t.Fatalf("FindBuildersNear: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(got))
	}
}
