-- +goose Up
ALTER TABLE professionals ADD COLUMN image_path TEXT;

-- Set images for the 9 new timber frame builders
UPDATE professionals SET image_path = 'builders/timber-frame-01.jpg' WHERE id = 9;
UPDATE professionals SET image_path = 'builders/timber-frame-02.jpg' WHERE id = 10;
UPDATE professionals SET image_path = 'builders/timber-detail.jpg' WHERE id = 11;
UPDATE professionals SET image_path = 'builders/red-barn.jpg' WHERE id = 12;
UPDATE professionals SET image_path = 'builders/palouse-barn.jpg' WHERE id = 13;
UPDATE professionals SET image_path = 'builders/greene-valley-barn.jpg' WHERE id = 14;
UPDATE professionals SET image_path = 'builders/colonial-house.jpg' WHERE id = 15;
UPDATE professionals SET image_path = 'builders/kenmore-mansion.jpg' WHERE id = 16;
UPDATE professionals SET image_path = 'builders/colonial-church.jpg' WHERE id = 17;

-- +goose Down
-- SQLite doesn't support DROP COLUMN directly, so we'd need to recreate the table
-- For simplicity, leaving this as a no-op since removing the column is rarely needed
