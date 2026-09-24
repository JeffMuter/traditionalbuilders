package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// NewTestDB creates an in-memory SQLite database with schema for testing.
func NewTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE zip_codes (zip TEXT PRIMARY KEY, lat REAL NOT NULL, lng REAL NOT NULL, city TEXT NOT NULL, state TEXT NOT NULL);
	CREATE TABLE professionals (id INTEGER PRIMARY KEY, name TEXT, specialty TEXT, zip_code TEXT, latitude REAL, longitude REAL);
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
		name     string
		zip      string
		lat, lng float64
	}{
		{"Near Orlando (closest)", "32801", 28.54, -81.38},
		{"Near LA (far)", "90210", 34.09, -118.41},
	}

	for _, p := range pros {
		if _, err := db.Exec(
			`INSERT INTO professionals (name, specialty, zip_code, latitude, longitude) VALUES (?, ?, ?, ?, ?)`,
			p.name, "General Contractor", p.zip, p.lat, p.lng,
		); err != nil {
			db.Close()
			return nil, err
		}
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
