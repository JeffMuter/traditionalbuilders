//go:build e2e

package e2e

import (
	"testing"
)

// TestLanding_LoadsAndSearches is the canonical user journey: land on the
// home page, type a zip, submit, and see results.
func TestLanding_LoadsAndSearches(t *testing.T) {
	page := newPage(t, "landing-loads-and-searches")
	page.page.MustNavigate(routeURL("/")).MustWaitLoad()

	expectVisible(t, page, "header", "/", "load")
	expectVisible(t, page, "footer", "/", "load")
	expectText(t, page, "h2", "Build with Timeless Beauty", "/", "load")

	builders := landingPage{session: page}.SearchZip("32801")
	expectURLContains(t, page, "/builders?zip=32801", "/", "submit zip search")
	expectVisible(t, page, `[data-testid="results"]`, "/builders", "show results")
	_ = builders
}

// TestLanding_HeroHeadingIsAccessibleNameable checks the page has a usable
// title, which axe and screen readers both rely on.
func TestLanding_HasTitle(t *testing.T) {
	page := newPage(t, "landing-has-title")
	page.page.MustNavigate(routeURL("/")).MustWaitLoad()

	title := page.page.MustEval(`() => document.title`).Str()
	if title == "" {
		t.Error("/: empty <title>")
	}
}
