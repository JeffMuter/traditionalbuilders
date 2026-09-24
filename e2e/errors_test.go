//go:build e2e

package e2e

import (
	"testing"
)

// TestErrors_NotFoundReturnsHome asserts an unmatched path shows the styled 404
// and its "Return Home" link actually returns the user to the landing page.
func TestErrors_NotFoundReturnsHome(t *testing.T) {
	page := newPage(t, "errors-not-found")
	page.page.MustNavigate(routeURL("/does-not-exist")).MustWaitLoad()

	expectVisible(t, page, `[data-testid="error-page"]`, "/does-not-exist", "404 body")
	expectText(t, page, `[data-testid="error-page"]`, "Page not found", "/does-not-exist", "404 body")

	page.page.MustElementR("a", "Return Home").MustClick()
	page.page.Timeout(shortWait).MustWaitLoad()
	expectVisible(t, page, `[data-testid="zip-form"]`, "/", "Return Home lands on search")
}
