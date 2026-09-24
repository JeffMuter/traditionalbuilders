-- +goose Up
CREATE TABLE IF NOT EXISTS zip_codes (
    zip   TEXT PRIMARY KEY,
    lat   REAL NOT NULL,
    lng   REAL NOT NULL,
    city  TEXT NOT NULL,
    state TEXT NOT NULL
);

-- The full ~41k-row US dataset is applied by internal/zipdata (embedded,
-- offline) via `go run ./cmd/seed-zips` / `server -seed-only`, not by a
-- migration. Data source: GeoNames postal codes, licensed CC BY 4.0
-- (https://www.geonames.org/). Earlier versions of this migration hard-coded
-- 113 bootstrap rows; that was removed once the full dataset became the
-- source of truth. See PLAN_ZIP_DATASET.md.

-- +goose Down
DROP TABLE IF EXISTS zip_codes;
