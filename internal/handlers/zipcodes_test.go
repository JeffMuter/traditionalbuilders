package handlers

import "testing"

func TestZipPattern(t *testing.T) {
	valid := []string{"32801", "90210", "32801-1234"}
	for _, z := range valid {
		if !zipPattern.MatchString(z) {
			t.Errorf("zipPattern rejected valid %q", z)
		}
	}
	invalid := []string{"", "1234", "123456", "abcde", "32801-12", "32801-12345"}
	for _, z := range invalid {
		if zipPattern.MatchString(z) {
			t.Errorf("zipPattern accepted invalid %q", z)
		}
	}
}

func TestBaseZip(t *testing.T) {
	cases := map[string]string{
		"32801":      "32801",
		"32801-1234": "32801",
		"":           "",
	}
	for in, want := range cases {
		if got := baseZip(in); got != want {
			t.Errorf("baseZip(%q) = %q, want %q", in, got, want)
		}
	}
}
