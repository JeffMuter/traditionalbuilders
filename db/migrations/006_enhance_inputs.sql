-- +goose Up
-- Provenance/audit columns for directory records.
ALTER TABLE professionals ADD COLUMN provider TEXT DEFAULT 'admin';
ALTER TABLE professionals ADD COLUMN verified_at DATE;

UPDATE professionals
SET provider = 'admin',
    verified_at = COALESCE(verified_at, CURRENT_DATE)
WHERE provider IS NULL;

-- Proximity search indexes.
-- zip_codes(zip) is already covered by its PRIMARY KEY, but the composite
-- lat/lng index helps bounding-box range scans as the table grows.
CREATE INDEX IF NOT EXISTS idx_prof_zip ON professionals(zip_code);
CREATE INDEX IF NOT EXISTS idx_prof_loc ON professionals(latitude, longitude);
CREATE INDEX IF NOT EXISTS idx_zip_codes_lat_lng ON zip_codes(lat, lng);

-- +goose Down
DROP INDEX IF EXISTS idx_zip_codes_lat_lng;
DROP INDEX IF EXISTS idx_prof_loc;
DROP INDEX IF EXISTS idx_prof_zip;
-- SQLite cannot DROP COLUMN; provider/verified_at are left in place on rollback.
