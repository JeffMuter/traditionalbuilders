package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emerald/traditionbuilders/internal/store"
)

// TestErrorCategorization tests the mapping from store errors to HTTP status codes.
func TestErrorCategorization(t *testing.T) {
	tests := []struct {
		name           string
		error          error
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "ErrZipNotFound returns 406",
			error:          store.ErrZipNotFound,
			expectedStatus: http.StatusNotAcceptable,
			expectedMsg:    "zip not found",
		},
		{
			name:           "DB error returns 500",
			error:          context.DeadlineExceeded,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal error",
		},
		{
			name:           "unknown error returns 500 (server fault)",
			error:          errors.New("some unexpected error"),
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeError(w, tt.error)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			var resp map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("response not JSON: %v", err)
			}

			if resp["error"] != tt.expectedMsg {
				t.Errorf("error = %q, want %q", resp["error"], tt.expectedMsg)
			}
		})
	}
}

// TestSearchBuilders_BadRequest returns 400 for invalid requests.
func TestSearchBuilders_BadRequest(t *testing.T) {
	h := &Handler{}

	tests := []struct {
		name string
		url  string
	}{
		{"non-numeric zip", "/api/builders?zip=abcde"},
		{"too short", "/api/builders?zip=1234"},
		{"too long", "/api/builders?zip=123456"},
		{"empty", "/api/builders?zip="},
		{"bad ZIP+4", "/api/builders?zip=12345-12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			req = req.WithContext(context.Background())

			w := httptest.NewRecorder()
			h.SearchBuilders(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}

			var resp map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("response not JSON: %v", err)
			}

			if resp["error"] != "invalid zip code" {
				t.Errorf("error = %q, want 'invalid zip code'", resp["error"])
			}
		})
	}
}
