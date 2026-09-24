-- +goose Up
-- Bookkeeping for vendored, embedded datasets applied outside the migration
-- chain (currently the GeoNames zip code dataset loaded by internal/zipdata).
-- Keyed by name + sha256 so a redeploy with a newer dataset self-heals, while
-- a restart with the same dataset is a single no-op SELECT.
CREATE TABLE IF NOT EXISTS data_seeds (
    name       TEXT PRIMARY KEY,
    sha256     TEXT NOT NULL,
    source     TEXT NOT NULL,
    row_count  INTEGER NOT NULL,
    loaded_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS data_seeds;
