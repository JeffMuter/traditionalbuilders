package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/emerald/traditionbuilders/internal/zipdata"
	_ "github.com/mattn/go-sqlite3"
)

// applyAllMigrations replays every db/migrations/*.sql Up section in order
// against db, reusing the helpers from migration_test.go (same package).
func applyAllMigrations(t *testing.T, db *sql.DB) {
	t.Helper()
	dir := migrationDir(t)
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no migration files found")
	}
	for _, f := range files {
		up, _ := splitGoose(t, f)
		applySQL(t, db, up)
	}
}

// TestFreshDB_FullZipDataset is the acceptance test for the zip dataset issue:
// on a clean database it runs the real migrations, applies the embedded
// dataset exactly as production does, and proves that a ZIP outside the old
// 113-row bootstrap set resolves through the proximity search.
func TestFreshDB_FullZipDataset(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	applyAllMigrations(t, db)

	// Migrations alone must not seed zip codes any more.
	var before int
	if err := db.QueryRow(`SELECT COUNT(*) FROM zip_codes`).Scan(&before); err != nil {
		t.Fatalf("count before: %v", err)
	}
	if before != 0 {
		t.Fatalf("zip_codes has %d rows after migrations; want 0", before)
	}

	if applied, err := zipdata.EnsureLoaded(db); err != nil {
		t.Fatalf("EnsureLoaded: %v", err)
	} else if !applied {
		t.Fatal("expected the dataset to be applied on a fresh DB")
	}

	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM zip_codes`).Scan(&n); err != nil {
		t.Fatalf("count after: %v", err)
	}
	if n < 40000 || n > 41490 {
		t.Fatalf("zip_codes has %d rows, want ~41k", n)
	}

	s := New(db)
	// 90210 is not in the old 113-row bootstrap set: a valid result (even if
	// empty) proves the full dataset resolved it; ErrZipNotFound would not.
	results, err := s.FindBuildersNear(context.Background(), "90210")
	if err != nil {
		t.Fatalf("FindBuildersNear(90210): %v", err)
	}
	if results == nil {
		t.Fatal("results must be non-nil")
	}

	// A genuinely unknown ZIP must still be reported as not found.
	if _, err := s.FindBuildersNear(context.Background(), "99999"); err != ErrZipNotFound {
		t.Fatalf("FindBuildersNear(99999) err = %v, want ErrZipNotFound", err)
	}
}
