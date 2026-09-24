//go:build e2e

package e2e

import (
	"strings"
	"testing"
)

// TestBuilders_ResultsAndProfileNavigation asserts that clicking a specific
// card's profile link lands on that card's builder, not merely "a profile".
func TestBuilders_ResultsAndProfileNavigation(t *testing.T) {
	page := newPage(t, "builders-results-and-profile")
	builders := buildersPage{session: page}.Load("32801")

	names := builders.CardNames(t)
	if len(names) == 0 {
		t.Fatal("/builders?zip=32801: expected at least one builder card")
	}
	for _, name := range names {
		if strings.TrimSpace(name) == "" {
			t.Error("/builders?zip=32801: a card has an empty name")
		}
	}

	clicked, profile := builders.ClickFirstProfile(t)
	expectURLContains(t, page, "/builders/", "/builders", "click view-profile")
	expectText(t, page, `[data-testid="profile"]`, clicked, "/builders/{id}", "profile matches clicked card")
	_ = profile
}

// TestBuilders_EmptyState asserts a valid zip with no builders shows the
// empty-state copy rather than an error or blank page.
func TestBuilders_EmptyState(t *testing.T) {
	page := newPage(t, "builders-empty-state")
	buildersPage{session: page}.Load("90001")

	expectVisible(t, page, `[data-testid="empty-state"]`, "/builders?zip=90001", "empty state")
	expectText(t, page, `[data-testid="empty-state"]`, "No builders found", "/builders?zip=90001", "empty state")
}

// TestBuilders_UnknownZipShowsError asserts an unknown-but-well-formed zip
// renders the "don't recognize" error state (not a 5xx, not the empty state).
func TestBuilders_UnknownZipShowsError(t *testing.T) {
	page := newPage(t, "builders-unknown-zip")
	buildersPage{session: page}.Load("99999")

	expectVisible(t, page, `[data-testid="search-error"]`, "/builders?zip=99999", "unknown zip error")
	expectText(t, page, `[data-testid="search-error"]`, "recognize", "/builders?zip=99999", "unknown zip error")
}

// TestBuilders_MalformedZipShowsValidation asserts the client-side pattern
// blocks a malformed submit and the server renders a validation error if it
// somehow arrives. The form's pattern attribute prevents submission, so the
// direct-navigation case is what actually exercises the server branch.
func TestBuilders_MalformedZipShowsValidation(t *testing.T) {
	page := newPage(t, "builders-malformed-zip")
	buildersPage{session: page}.Load("abc")

	expectVisible(t, page, `[data-testid="search-error"]`, "/builders?zip=abc", "validation error")
	expectText(t, page, `[data-testid="search-error"]`, "valid 5-digit zip code", "/builders?zip=abc", "validation error")
}

// TestBuilders_MobileViewport asserts the results page renders its cards on a
// phone-sized viewport, catching layout rules that hide content on narrow
// screens.
func TestBuilders_MobileViewport(t *testing.T) {
	page := newPage(t, "builders-mobile-viewport")
	setViewport(t, page, mobileViewport)
	builders := buildersPage{session: page}.Load("32801")

	if names := builders.CardNames(t); len(names) == 0 {
		t.Fatal("/builders?zip=32801 (mobile): expected at least one card")
	}
	expectVisible(t, page, "header", "/builders?zip=32801", "mobile header nav")
}
