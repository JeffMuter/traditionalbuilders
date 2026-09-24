package store

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func newTestStore(t *testing.T) *Store {
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

	zips := []struct {
		zip      string
		lat, lng float64
		city, st string
	}{
		{"32801", 28.5383, -81.3792, "Orlando", "FL"},
		{"29401", 32.7765, -79.9311, "Charleston", "SC"},
		{"90210", 34.0901, -118.4065, "Beverly Hills", "CA"},
	}
	for _, z := range zips {
		if _, err := db.Exec(
			`INSERT INTO zip_codes (zip, lat, lng, city, state) VALUES (?, ?, ?, ?, ?)`,
			z.zip, z.lat, z.lng, z.city, z.st,
		); err != nil {
			t.Fatalf("seed zip: %v", err)
		}
	}

	pros := []struct {
		name     string
		zip      string
		lat, lng float64
	}{
		{"Near Orlando", "32801", 28.54, -81.38},
		{"Near Charleston", "29401", 32.78, -79.93},
		{"Near LA", "90210", 34.09, -118.41},
	}
	for i, p := range pros {
		if _, err := db.Exec(
			`INSERT INTO professionals (id, name, specialty, zip_code, latitude, longitude) VALUES (?, ?, 'test', ?, ?, ?)`,
			i+1, p.name, p.zip, p.lat, p.lng,
		); err != nil {
			t.Fatalf("seed pro: %v", err)
		}
	}

	return New(db)
}

func TestFindBuildersNear_OrdersByDistance(t *testing.T) {
	s := newTestStore(t)

	got, err := s.FindBuildersNear(context.Background(), "32801")
	if err != nil {
		t.Fatalf("FindBuildersNear: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected at least one result")
	}
	if got[0].Name != "Near Orlando" {
		t.Errorf("closest = %q, want Near Orlando", got[0].Name)
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].DistanceMiles > got[i].DistanceMiles {
			t.Errorf("results not sorted: %v > %v", got[i-1].DistanceMiles, got[i].DistanceMiles)
		}
	}
}

func TestFindBuildersNear_UnknownZip(t *testing.T) {
	s := newTestStore(t)

	got, err := s.FindBuildersNear(context.Background(), "00000")
	if err != ErrZipNotFound {
		t.Fatalf("err = %v, want ErrZipNotFound", err)
	}
	if got == nil {
		t.Error("expected non-nil slice")
	}
}

func TestFindBuildersNear_BoundingBoxExcludesFarBuilders(t *testing.T) {
	s := newTestStore(t)

	got, err := s.FindBuildersNear(context.Background(), "32801")
	if err != nil {
		t.Fatalf("FindBuildersNear: %v", err)
	}
	for _, b := range got {
		if b.Name == "Near LA" {
			t.Error("Los Angeles builder should be excluded by the 500-mile bounding box")
		}
	}
}

func TestHaversine(t *testing.T) {
	// Orlando -> Charleston is roughly 350 miles; allow a wide tolerance.
	dist := haversine(28.5383, -81.3792, 32.7765, -79.9311)
	if dist < 300 || dist > 400 {
		t.Errorf("haversine = %.1f, want ~350", dist)
	}
}
