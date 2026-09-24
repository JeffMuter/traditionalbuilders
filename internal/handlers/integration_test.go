package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/emerald/traditionbuilders/internal/models"
	"github.com/emerald/traditionbuilders/internal/store"
	_ "github.com/mattn/go-sqlite3"
)

// setupIntegrationTest creates a test server with fresh database.
func setupIntegrationTest(t *testing.T) *Handler {
	t.Helper()

	db, err := NewTestDB()
	if err != nil {
		t.Fatalf("setup db: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	return &Handler{Store: store.New(db)}
}

// setupTestServer builds a full HTTP server (routes + middleware + static file
// server) backed by a fresh in-memory DB. It returns the server and its base
// URL. Tests that need real routing, the FileServer, or middleware should use
// this instead of calling handlers directly so they exercise the same mux as
// production.
//
// The static directory is resolved relative to this package's source tree so
// the FileServer finds the checked-in assets regardless of the test's working
// directory.
func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	db, err := NewTestDB()
	if err != nil {
		t.Fatalf("setup db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	h := &Handler{Store: store.New(db)}
	server := httptest.NewServer(h.Routes(db, repoStaticDir(t)))
	t.Cleanup(server.Close)
	return server
}

// repoStaticDir returns the absolute path to the static/ directory in the
// repository root, walking up from the handlers package directory.
func repoStaticDir(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", "static"))
	if err != nil {
		t.Fatalf("resolve static dir: %v", err)
	}
	return abs
}

// NewTestDB creates an in-memory SQLite database matching the production
// schema end-state (after all migrations) and seeds a small deterministic
// fixture set. It is shared by handler tests, the render-contract tests, and
// the browser E2E suite so all three assert against the same data.
func NewTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE zip_codes (zip TEXT PRIMARY KEY, lat REAL NOT NULL, lng REAL NOT NULL, city TEXT NOT NULL, state TEXT NOT NULL);
	CREATE TABLE professionals (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		phone TEXT,
		specialty TEXT NOT NULL,
		location TEXT,
		bio TEXT,
		zip_code TEXT,
		latitude REAL,
		longitude REAL,
		image_path TEXT,
		provider TEXT DEFAULT 'admin',
		verified_at DATE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE projects (
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
	`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}

	// Seed test data: zip codes and professionals
	zips := []struct {
		zip      string
		lat, lng float64
		city, st string
	}{
		{"32801", 28.5383, -81.3792, "Orlando", "FL"},
		{"90210", 34.0901, -118.4065, "Beverly Hills", "CA"},
		{"32801-1234", 28.5383, -81.3792, "Orlando", "FL"}, // ZIP+4 is valid
		{"10001", 40.7506, -73.9971, "New York", "NY"},     // valid zip with no builders
	}

	for _, z := range zips {
		if _, err := db.Exec(
			`INSERT INTO zip_codes (zip, lat, lng, city, state) VALUES (?, ?, ?, ?, ?)`,
			z.zip, z.lat, z.lng, z.city, z.st,
		); err != nil {
			db.Close()
			return nil, err
		}
	}

	pros := []struct {
		name  string
		email string
		zip   string
		lat   float64
		lng   float64
	}{
		{"Near Orlando (closest)", "orlando@example.com", "32801", 28.54, -81.38},
		{"Near LA (far)", "la@example.com", "90210", 34.09, -118.41},
	}

	for _, p := range pros {
		if _, err := db.Exec(
			`INSERT INTO professionals (name, email, phone, specialty, location, bio, zip_code, latitude, longitude)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			p.name,
			p.email,
			"(407) 555-0101",
			"General Contractor",
			"Test Location",
			"A test builder biography.",
			p.zip, p.lat, p.lng,
		); err != nil {
			db.Close()
			return nil, err
		}
	}

	// Seed one project for the first professional so profile pages exercise the
	// projects section, and one image path for the image branch.
	if _, err := db.Exec(
		`INSERT INTO projects (professional_id, title, description, location, cost_estimate, completed_at)
		 VALUES (1, 'Restored Farmhouse', 'A careful restoration.', 'Orlando, FL', 250000, '2023-06-01')`,
	); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// TestSearchBuilders_Integration tests the full HTTP request pipeline.
func TestSearchBuilders_Integration(t *testing.T) {
	h := setupIntegrationTest(t)

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		checkResponse  func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "valid zip returns JSON list",
			url:            "/api/builders?zip=32801",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				if resp.Code != http.StatusOK {
					t.Errorf("status = %d, want %d", resp.Code, http.StatusOK)
				}

				var results []models.BuilderResult
				if err := json.Unmarshal(resp.Body.Bytes(), &results); err != nil {
					t.Fatalf("response not JSON: %v", err)
				}

				if len(results) == 0 {
					t.Error("expected at least one result")
				}

				// Verify results are sorted by distance
				for i := 1; i < len(results); i++ {
					if results[i-1].DistanceMiles > results[i].DistanceMiles {
						t.Errorf("results not sorted: %v > %v",
							results[i-1].DistanceMiles,
							results[i].DistanceMiles)
					}
				}
			},
		},
		{
			name:           "invalid zip format returns 400",
			url:            "/api/builders?zip=abcde",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				if resp.Code != http.StatusBadRequest {
					t.Errorf("status = %d, want %d", resp.Code, http.StatusBadRequest)
				}

				var respJSON map[string]string
				if err := json.Unmarshal(resp.Body.Bytes(), &respJSON); err != nil {
					t.Fatalf("response not JSON: %v", err)
				}

				if respJSON["error"] != "invalid zip code" {
					t.Errorf("error = %q, want 'invalid zip code'", respJSON["error"])
				}
			},
		},
		{
			name:           "empty zip returns 400",
			url:            "/api/builders?zip=",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil, // Just verify status
		},
		{
			name:           "ZIP+4 format is accepted",
			url:            "/api/builders?zip=32801-1234",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				if resp.Code != http.StatusOK {
					t.Errorf("status = %d, want %d", resp.Code, http.StatusOK)
				}

				var results []models.BuilderResult
				if err := json.Unmarshal(resp.Body.Bytes(), &results); err != nil {
					t.Fatalf("response not JSON: %v", err)
				}

				// ZIP+4 should be normalized to 5 digits
				if len(results) == 0 {
					t.Error("expected results after ZIP+4 normalization")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			req = req.WithContext(context.Background())

			rec := httptest.NewRecorder()
			h.SearchBuilders(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.expectedStatus)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

// TestJSON_ContentType verifies the API contract: JSON content-type and a
// non-null JSON array body.
//
// NOTE: the plan called for a CORS assertion here, but the app sets no CORS
// headers and the htmx frontend calls this API same-origin, so CORS is not
// part of the current contract. See report.
func TestJSON_ContentType(t *testing.T) {
	h := setupIntegrationTest(t)

	req, _ := http.NewRequest("GET", "/api/builders?zip=32801", nil)
	req = req.WithContext(context.Background())

	rec := httptest.NewRecorder()
	h.SearchBuilders(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	body := rec.Body.String()
	if body == "null\n" || body == "null" {
		t.Error("body is JSON null; want a non-null array")
	}
}
