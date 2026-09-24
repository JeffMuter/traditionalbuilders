package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestBuildersPage_ValidZip renders the HTML page and checks the search results.
func TestBuildersPage_ValidZip(t *testing.T) {
	h := setupIntegrationTest(t)

	req, _ := http.NewRequest("GET", "/builders?zip=32801", nil)
	req = req.WithContext(context.Background())

	rec := httptest.NewRecorder()
	h.BuildersPage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Traditional Builders Near 32801") {
		t.Error("expected results heading with zip")
	}
	if !strings.Contains(body, "Near Orlando (closest)") {
		t.Error("expected nearest builder name in HTML")
	}
	if !strings.Contains(body, "htmx.org") {
		t.Error("expected htmx script in page")
	}
}

// TestBuildersPage_InvalidZip_ShowsError verifies the input is re-populated
// with a validation message.
func TestBuildersPage_InvalidZip_ShowsError(t *testing.T) {
	h := setupIntegrationTest(t)

	req, _ := http.NewRequest("GET", "/builders?zip=abc", nil)
	req = req.WithContext(context.Background())

	rec := httptest.NewRecorder()
	h.BuildersPage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "valid 5-digit zip code") {
		t.Error("expected validation error message")
	}
}

// TestBuildersPage_UnknownZip_ShowsNotFound verifies a valid-format unknown zip
// renders the "don't recognize" message rather than a 5xx.
func TestBuildersPage_UnknownZip_ShowsNotFound(t *testing.T) {
	h := setupIntegrationTest(t)

	req, _ := http.NewRequest("GET", "/builders?zip=99999", nil)
	req = req.WithContext(context.Background())

	rec := httptest.NewRecorder()
	h.BuildersPage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "recognize") {
		t.Error("expected zip-not-found message")
	}
}
