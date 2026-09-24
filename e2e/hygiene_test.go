//go:build e2e

package e2e

import (
	"testing"
)

// hygieneRoutes are the pages every navigation must keep clean: no console
// errors and no broken same-origin assets.
var hygieneRoutes = map[string]string{
	"/":                   "landing",
	"/gallery":            "gallery",
	"/builders?zip=32801": "builders-results",
	"/builders?zip=90001": "builders-empty",
	"/builders?zip=99999": "builders-error",
	"/builders/1":         "profile",
	"/builders/9999":      "profile-not-found",
	"/does-not-exist":     "error-404",
}

// TestHygiene_NoConsoleErrorsOrBrokenAssets walks every route and fails on a
// console error or a same-origin request ≥400. External asset failures are
// warnings only (see assertions.go).
func TestHygiene_NoConsoleErrorsOrBrokenAssets(t *testing.T) {
	for route, name := range hygieneRoutes {
		t.Run(name, func(t *testing.T) {
			page := newPage(t, "hygiene-"+name)
			page.page.MustNavigate(routeURL(route)).MustWaitLoad()

			expectNoConsoleErrors(t, page, route, "navigate")
			expectNoBrokenSameOriginAssets(t, page, route, "navigate")
		})
	}
}
