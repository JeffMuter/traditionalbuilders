package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestAPIContract_SearchBuilders pins the /api/builders JSON contract end-to-end
// through the real mux (not the handler in isolation), so routing and middleware
// are covered too.
func TestAPIContract_SearchBuilders(t *testing.T) {
	server := setupTestServer(t)

	t.Run("valid zip returns JSON array", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/builders?zip=32801")
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		body, _ := io.ReadAll(resp.Body)
		if strings.TrimSpace(string(body)) == "null" {
			t.Fatal("body is JSON null; want a non-nil array")
		}

		var results []struct {
			ID            int     `json:"ID"`
			Name          string  `json:"Name"`
			DistanceMiles float64 `json:"DistanceMiles"`
		}
		if err := json.Unmarshal(body, &results); err != nil {
			t.Fatalf("body is not a JSON array: %v (%s)", err, body)
		}
		if len(results) == 0 {
			t.Fatal("expected at least one result")
		}
		for i := 1; i < len(results); i++ {
			if results[i-1].DistanceMiles > results[i].DistanceMiles {
				t.Errorf("results not sorted by distance: %v > %v",
					results[i-1].DistanceMiles, results[i].DistanceMiles)
			}
		}
	})

	t.Run("malformed zip returns 400", func(t *testing.T) {
		for _, zip := range []string{"abcde", "1234", "123456", ""} {
			resp, err := http.Get(server.URL + "/api/builders?zip=" + zip)
			if err != nil {
				t.Fatalf("GET zip=%q: %v", zip, err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("zip=%q: status = %d, want 400", zip, resp.StatusCode)
			}
		}
	})

	t.Run("unknown zip returns 406", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/builders?zip=99999")
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotAcceptable {
			t.Fatalf("status = %d, want 406", resp.StatusCode)
		}
		var body map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("body not JSON object: %v", err)
		}
		if body["error"] != "zip not found" {
			t.Errorf("error = %q, want 'zip not found'", body["error"])
		}
	})
}
