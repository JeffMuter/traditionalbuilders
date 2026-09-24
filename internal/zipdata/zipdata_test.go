package zipdata

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestMetadata_MatchesEmbeddedFile verifies dataset.json parses and its
// declared row_count is consistent with the embedded TSV after filtering.
func TestMetadata_MatchesEmbeddedFile(t *testing.T) {
	m, err := Metadata()
	if err != nil {
		t.Fatalf("Metadata: %v", err)
	}
	if m.Name != DatasetName {
		t.Errorf("Name = %q, want %q", m.Name, DatasetName)
	}
	if m.SHA256 == "" {
		t.Error("SHA256 is empty")
	}
	if m.License == "" || m.Attribution == "" {
		t.Error("license/attribution must be recorded (CC BY 4.0 requirement)")
	}
	if m.RowCount < 40000 || m.RowCount > 45000 {
		t.Errorf("RowCount = %d, want ~41k", m.RowCount)
	}

	got, err := Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if got != m.RowCount {
		t.Errorf("filtered rows = %d, dataset.json says %d", got, m.RowCount)
	}
}

// TestEach_Integrity verifies every emitted row is a unique, well-formed,
// in-range 5-digit ZIP.
func TestEach_Integrity(t *testing.T) {
	seen := make(map[string]bool)
	if err := Each(func(r Row) error {
		if len(r.Zip) != 5 || !allDigits(r.Zip) {
			t.Errorf("bad zip %q", r.Zip)
		}
		if seen[r.Zip] {
			t.Errorf("duplicate zip %q", r.Zip)
		}
		seen[r.Zip] = true
		if r.Lat < minLat || r.Lat > maxLat || r.Lng < minLng || r.Lng > maxLng {
			t.Errorf("zip %s out of bounds: %v,%v", r.Zip, r.Lat, r.Lng)
		}
		if r.City == "" || r.State == "" {
			t.Errorf("zip %s missing city/state", r.Zip)
		}
		return nil
	}); err != nil {
		t.Fatalf("Each: %v", err)
	}
	if len(seen) < 40000 {
		t.Fatalf("only %d unique zips", len(seen))
	}
	// A zip outside the original 113-row bootstrap set must exist.
	if !seen["90210"] {
		t.Error("expected 90210 (Beverly Hills) in dataset")
	}
}

// TestParseLine rejects malformed input and accepts a valid row.
func TestParseLine(t *testing.T) {
	cases := []struct {
		in  string
		ok  bool
		zip string
	}{
		{"32801\t28.5383\t-81.3792\tOrlando\tFL", true, "32801"},
		{"90210\t34.0901\t-118.4065\tBeverly Hills\tCA", true, "90210"},
		{"32801-1234\t28.5\t-81.3\tOrlando\tFL", false, ""},  // ZIP+4
		{"ABCDE\t28.5\t-81.3\tNowhere\tFL", false, ""},       // non-numeric
		{"99999\t999.0\t-81.3\tBad\tFL", false, ""},          // lat OOR
		{"99998\t28.5\t10.0\tBad\tFL", false, ""},            // lng OOR
		{"99997\t28.5\t-81.3\tTooFew\tFL\tx", true, "99997"}, // extra cols ignored
		{"99996\tnotafloat\t-81.3\tBad\tFL", false, ""},      // unparseable
		{"too\tfew", false, ""},                              // too few
	}
	for _, c := range cases {
		row, ok := parseLine(c.in)
		if ok != c.ok {
			t.Errorf("parseLine(%q) ok = %v, want %v", c.in, ok, c.ok)
			continue
		}
		if ok && row.Zip != c.zip {
			t.Errorf("parseLine(%q) zip = %q, want %q", c.in, row.Zip, c.zip)
		}
	}
}

func newSeedDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	stmts := []string{
		`CREATE TABLE zip_codes (zip TEXT PRIMARY KEY, lat REAL NOT NULL, lng REAL NOT NULL, city TEXT NOT NULL, state TEXT NOT NULL)`,
		`CREATE TABLE data_seeds (name TEXT PRIMARY KEY, sha256 TEXT NOT NULL, source TEXT NOT NULL, row_count INTEGER NOT NULL, loaded_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("exec %q: %v", s, err)
		}
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// TestEnsureLoaded_Idempotent verifies the first call loads and records, and
// the second is a no-op.
func TestEnsureLoaded_Idempotent(t *testing.T) {
	db := newSeedDB(t)

	applied, err := EnsureLoaded(db)
	if err != nil {
		t.Fatalf("first EnsureLoaded: %v", err)
	}
	if !applied {
		t.Fatal("first call should apply the dataset")
	}

	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM zip_codes`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n < 40000 {
		t.Fatalf("zip_codes has %d rows, want ~41k", n)
	}

	var seeds int
	if err := db.QueryRow(`SELECT COUNT(*) FROM data_seeds WHERE name = ?`, DatasetName).Scan(&seeds); err != nil {
		t.Fatalf("data_seeds: %v", err)
	}
	if seeds != 1 {
		t.Fatalf("data_seeds rows = %d, want 1", seeds)
	}

	applied, err = EnsureLoaded(db)
	if err != nil {
		t.Fatalf("second EnsureLoaded: %v", err)
	}
	if applied {
		t.Error("second call should be a no-op")
	}
}

// TestEnsureLoaded_ReappliesOnVersionChange verifies a stale sha triggers a
// reload that replaces the rows.
func TestEnsureLoaded_ReappliesOnVersionChange(t *testing.T) {
	db := newSeedDB(t)
	if _, err := EnsureLoaded(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Corrupt the stored sha and add a junk row.
	if _, err := db.Exec(`UPDATE data_seeds SET sha256 = 'stale' WHERE name = ?`, DatasetName); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO zip_codes (zip, lat, lng, city, state) VALUES ('00000', 40, -100, 'Junk', 'XX')`); err != nil {
		t.Fatalf("insert junk: %v", err)
	}

	applied, err := EnsureLoaded(db)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if !applied {
		t.Fatal("stale sha should trigger a reload")
	}
	var junk int
	if err := db.QueryRow(`SELECT COUNT(*) FROM zip_codes WHERE zip = '00000'`).Scan(&junk); err != nil {
		t.Fatalf("count junk: %v", err)
	}
	if junk != 0 {
		t.Error("junk row survived reload; DELETE did not run")
	}
}

// TestEnsureLoaded_MissingTableIsError verifies a clear error when migrations
// have not run, so callers can log-and-continue.
func TestEnsureLoaded_MissingTableIsError(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if _, err := EnsureLoaded(db); err == nil {
		t.Error("expected error when zip_codes/data_seeds are absent")
	}
}
