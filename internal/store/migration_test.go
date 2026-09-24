package store

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// migrationDir locates db/migrations from the package directory.
func migrationDir(t *testing.T) string {
	t.Helper()
	// internal/store -> repo root
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}
	dir := filepath.Join(root, "db", "migrations")
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("migrations dir: %v", err)
	}
	return dir
}

// splitGoose returns the Up and Down SQL from a goose migration file.
func splitGoose(t *testing.T, path string) (up, down string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	content := string(raw)
	upIdx := strings.Index(content, "-- +goose Up")
	downIdx := strings.Index(content, "-- +goose Down")
	if upIdx < 0 {
		t.Fatalf("%s: missing +goose Up", path)
	}
	upBody := content[upIdx+len("-- +goose Up"):]
	if downIdx < 0 {
		return strings.TrimSpace(upBody), ""
	}
	return strings.TrimSpace(content[upIdx+len("-- +goose Up") : downIdx]),
		strings.TrimSpace(content[downIdx+len("-- +goose Down"):])
}

// applySQL runs a possibly multi-statement SQL blob.
func applySQL(t *testing.T, db *sql.DB, sqlText string) {
	t.Helper()
	if strings.TrimSpace(sqlText) == "" {
		return
	}
	if _, err := db.Exec(sqlText); err != nil {
		t.Fatalf("exec SQL: %v\nSQL:\n%s", err, sqlText)
	}
}

// TestMigration006_AppliesCleanly applies all migrations in order against a
// fresh in-memory database, then verifies migration 006's effects.
func TestMigration006_AppliesCleanly(t *testing.T) {
	dir := migrationDir(t)
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no migration files found")
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	for _, f := range files {
		up, _ := splitGoose(t, f)
		applySQL(t, db, up)
	}

	// 006 adds provider/verified_at to professionals.
	rows, err := db.Query("PRAGMA table_info(professionals)")
	if err != nil {
		t.Fatalf("pragma: %v", err)
	}
	defer rows.Close()
	cols := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan pragma: %v", err)
		}
		cols[name] = true
	}
	for _, want := range []string{"provider", "verified_at"} {
		if !cols[want] {
			t.Errorf("professionals missing column %q after migration 006", want)
		}
	}

	// Indexes from 006 should exist.
	idxRows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='index'`)
	if err != nil {
		t.Fatalf("index query: %v", err)
	}
	defer idxRows.Close()
	indexes := map[string]bool{}
	for idxRows.Next() {
		var name string
		if err := idxRows.Scan(&name); err != nil {
			t.Fatalf("scan index: %v", err)
		}
		indexes[name] = true
	}
	for _, want := range []string{"idx_prof_zip", "idx_prof_loc", "idx_zip_codes_lat_lng"} {
		if !indexes[want] {
			t.Errorf("missing index %q after migration 006", want)
		}
	}
}

// TestMigration006_RollbackDropsIndexes verifies the Down section removes the
// indexes added by 006. (SQLite cannot DROP COLUMN, so columns remain.)
func TestMigration006_RollbackDropsIndexes(t *testing.T) {
	dir := migrationDir(t)
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var mig006 string
	for _, f := range files {
		up, _ := splitGoose(t, f)
		applySQL(t, db, up)
		if strings.Contains(filepath.Base(f), "006") {
			mig006 = f
		}
	}
	if mig006 == "" {
		t.Fatal("migration 006 not found")
	}

	_, down := splitGoose(t, mig006)
	applySQL(t, db, down)

	idxRows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='index'`)
	if err != nil {
		t.Fatalf("index query: %v", err)
	}
	defer idxRows.Close()
	indexes := map[string]bool{}
	for idxRows.Next() {
		var name string
		if err := idxRows.Scan(&name); err != nil {
			t.Fatalf("scan index: %v", err)
		}
		indexes[name] = true
	}
	for _, gone := range []string{"idx_prof_zip", "idx_prof_loc", "idx_zip_codes_lat_lng"} {
		if indexes[gone] {
			t.Errorf("index %q still present after rollback", gone)
		}
	}
}
