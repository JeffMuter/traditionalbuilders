-- +goose Up
CREATE TABLE IF NOT EXISTS professionals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    phone TEXT,
    specialty TEXT NOT NULL,
    location TEXT,
    bio TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    professional_id INTEGER,
    title TEXT NOT NULL,
    description TEXT,
    location TEXT,
    cost_estimate INTEGER,
    completed_at DATE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (professional_id) REFERENCES professionals(id)
);

-- +goose Down
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS professionals;
