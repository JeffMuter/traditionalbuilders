package main

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

// buildTestArchive creates an in-memory US.zip containing a US.txt with the
// given tab-separated lines.
func buildTestArchive(t *testing.T, lines []string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("US.txt")
	if err != nil {
		t.Fatalf("create US.txt: %v", err)
	}
	if _, err := w.Write([]byte(strings.Join(lines, "\n") + "\n")); err != nil {
		t.Fatalf("write US.txt: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

// TestParseGeonames_ValidRows verifies well-formed GeoNames rows parse correctly.
func TestParseGeonames_ValidRows(t *testing.T) {
	lines := []string{
		"US\t32801\tOrlando\tFlorida\tFL\tOrange\t\t\t\t28.5383\t-81.3792\t1",
		"US\t90210\tBeverly Hills\tCalifornia\tCA\tLos Angeles\t\t\t\t34.0901\t-118.4065\t1",
		"US\t29401\tCharleston\tSouth Carolina\tSC\tCharleston\t\t\t\t32.7765\t-79.9311\t1",
	}
	rows, err := parseGeonames(buildTestArchive(t, lines))
	if err != nil {
		t.Fatalf("parseGeonames: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("len(rows) = %d, want 3", len(rows))
	}
	if rows[0].Zip != "32801" || rows[0].City != "Orlando" || rows[0].State != "FL" {
		t.Errorf("row[0] = %+v", rows[0])
	}
	if rows[0].Lat != 28.5383 || rows[0].Lng != -81.3792 {
		t.Errorf("row[0] coords = %v,%v", rows[0].Lat, rows[0].Lng)
	}
}

// TestParseGeonames_SkipsMalformed verifies the integrity skiplist rejects
// ZIP+4 codes, bad coordinates, and out-of-range rows.
func TestParseGeonames_SkipsMalformed(t *testing.T) {
	lines := []string{
		"US\t32801\tOrlando\tFlorida\tFL\tOrange\t\t\t\t28.5383\t-81.3792\t1",       // valid
		"US\t32801-1234\tOrlando\tFlorida\tFL\tOrange\t\t\t\t28.5383\t-81.3792\t1",  // ZIP+4 rejected
		"US\tABCDE\tNowhere\tFlorida\tFL\tOrange\t\t\t\t28.5383\t-81.3792\t1",       // non-numeric rejected
		"US\t99999\tBad Lat\tFlorida\tFL\tOrange\t\t\t\t999.0\t-81.3792\t1",         // lat out of range
		"US\t99998\tBad Lng\tFlorida\tFL\tOrange\t\t\t\t28.5383\t10.0\t1",           // lng out of range
		"US\t99997\tToo Few\tFlorida\tFL",                                           // too few columns
		"US\t99996\tBad Float\tFlorida\tFL\tOrange\t\t\t\tnot-a-float\t-81.3792\t1", // unparseable lat
	}
	rows, err := parseGeonames(buildTestArchive(t, lines))
	if err != nil {
		t.Fatalf("parseGeonames: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1 (only the valid row)", len(rows))
	}
	if rows[0].Zip != "32801" {
		t.Errorf("surviving row = %+v, want 32801", rows[0])
	}
}

// TestParseGeonames_MissingUSFile verifies a clear error when US.txt is absent.
func TestParseGeonames_MissingUSFile(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if _, err := zw.Create("README.txt"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	_, err := parseGeonames(buf.Bytes())
	if err == nil {
		t.Fatal("expected error for missing US.txt")
	}
	if !strings.Contains(err.Error(), "US.txt") {
		t.Errorf("error = %v, want mention of US.txt", err)
	}
}

// TestParseGeonames_AllFiftyStates verifies every US state abbreviation can be
// represented in the parsed output (data-integrity coverage check).
func TestParseGeonames_AllFiftyStates(t *testing.T) {
	states := []string{
		"AL", "AK", "AZ", "AR", "CA", "CO", "CT", "DE", "FL", "GA",
		"HI", "ID", "IL", "IN", "IA", "KS", "KY", "LA", "ME", "MD",
		"MA", "MI", "MN", "MS", "MO", "MT", "NE", "NV", "NH", "NJ",
		"NM", "NY", "NC", "ND", "OH", "OK", "OR", "PA", "RI", "SC",
		"SD", "TN", "TX", "UT", "VT", "VA", "WA", "WV", "WI", "WY",
	}

	lines := make([]string, 0, len(states))
	for i, st := range states {
		zip := 10000 + i // synthetic but 5-digit
		lines = append(lines, strings.Join([]string{
			"US", itoa(zip), "City" + st, "State " + st, st, "County", "", "", "",
			"38.0", "-97.0", "1",
		}, "\t"))
	}

	rows, err := parseGeonames(buildTestArchive(t, lines))
	if err != nil {
		t.Fatalf("parseGeonames: %v", err)
	}
	seen := map[string]bool{}
	for _, r := range rows {
		seen[r.State] = true
	}
	for _, st := range states {
		if !seen[st] {
			t.Errorf("state %s missing from parsed rows", st)
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [8]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
