-- +goose Up
ALTER TABLE professionals ADD COLUMN zip_code TEXT;
ALTER TABLE professionals ADD COLUMN latitude REAL;
ALTER TABLE professionals ADD COLUMN longitude REAL;

-- +goose Down
-- SQLite does not support DROP COLUMN; recreate the table to reverse if needed
