//go:build e2e

package e2e

import (
	"strings"
	"testing"
)

// TestProfile_KnownBuilder asserts a known profile renders the name, contact
// buttons, a well-formed mailto: link, and the projects section.
func TestProfile_KnownBuilder(t *testing.T) {
	page := newPage(t, "profile-known-builder")
	profilePage{session: page}.Load("1")

	expectVisible(t, page, `[data-testid="profile"]`, "/builders/1", "profile renders")
	expectVisible(t, page, `[data-testid="projects"]`, "/builders/1", "projects section")

	emailLink := page.page.MustElement(`a[href^="mailto:"]`)
	href, err := emailLink.Attribute("href")
	if err != nil || href == nil {
		t.Fatal("/builders/1: could not read mailto href")
	}
	if !strings.HasPrefix(*href, "mailto:") || strings.TrimPrefix(*href, "mailto:") == "" {
		t.Errorf("/builders/1: malformed mailto href %q", *href)
	}
}

// TestProfile_UnknownAndMalformedID asserts both a numeric-but-unknown id and a
// non-numeric id render the styled not-found body. The status code is verified
// separately by the render-contract tests; here we assert the user-visible page.
func TestProfile_UnknownAndMalformedID(t *testing.T) {
	for _, tc := range []struct{ name, id string }{
		{"unknown", "9999"},
		{"non-numeric", "abc"},
		{"zero", "0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			page := newPage(t, "profile-not-found-"+tc.name)
			profilePage{session: page}.Load(tc.id)

			expectVisible(t, page, `[data-testid="profile-not-found"]`, "/builders/"+tc.id, "not-found body")
			expectText(t, page, `[data-testid="profile-not-found"]`,
				"couldn't find that builder", "/builders/"+tc.id, "not-found body")
		})
	}
}

// TestProfile_SearchOtherBuilders asserts the "Search Other Builders" control
// returns the user to the search page.
func TestProfile_SearchOtherBuilders(t *testing.T) {
	page := newPage(t, "profile-search-other")
	profilePage{session: page}.Load("1")

	page.page.MustElement(`[data-testid="search-other"]`).MustClick()
	page.page.Timeout(shortWait).MustWaitLoad()
	expectVisible(t, page, `[data-testid="zip-form"]`, "/", "search-other returns home")
}
