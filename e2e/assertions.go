//go:build e2e

package e2e

import (
	"strings"
	"testing"
)

// This file is the single assertion vocabulary for the suite. Every failure
// names the route and the user step so an agent (or human) can act on it
// without re-reading the test body. See PLAN_FRONTEND_TESTING.md § conventions.

// expectVisible asserts an element matching selector exists, failing with the
// route and step in the message.
func expectVisible(t *testing.T, page *pageSession, selector, route, step string) {
	t.Helper()
	if _, err := page.page.Timeout(shortWait).Element(selector); err != nil {
		t.Errorf("%s [%s]: expected visible %q: %v", route, step, selector, err)
	}
}

// expectText asserts the element's text contains want.
func expectText(t *testing.T, page *pageSession, selector, want, route, step string) {
	t.Helper()
	el, err := page.page.Timeout(shortWait).Element(selector)
	if err != nil {
		t.Errorf("%s [%s]: element %q not found: %v", route, step, selector, err)
		return
	}
	got, err := el.Text()
	if err != nil {
		t.Errorf("%s [%s]: could not read text of %q: %v", route, step, selector, err)
		return
	}
	if !strings.Contains(got, want) {
		t.Errorf("%s [%s]: %q text = %q, want it to contain %q", route, step, selector, got, want)
	}
}

// expectURLContains asserts the current URL contains the fragment.
func expectURLContains(t *testing.T, page *pageSession, fragment, route, step string) {
	t.Helper()
	info, err := page.page.Info()
	if err != nil {
		t.Errorf("%s [%s]: could not read page info: %v", route, step, err)
		return
	}
	if !strings.Contains(info.URL, fragment) {
		t.Errorf("%s [%s]: url = %q, want it to contain %q", route, step, info.URL, fragment)
	}
}

// expectAttr asserts the named attribute of the element equals want.
func expectAttr(t *testing.T, page *pageSession, selector, attr, want, route, step string) {
	t.Helper()
	el, err := page.page.Timeout(shortWait).Element(selector)
	if err != nil {
		t.Errorf("%s [%s]: element %q not found: %v", route, step, selector, err)
		return
	}
	got, err := el.Attribute(attr)
	if err != nil {
		t.Errorf("%s [%s]: attribute %q on %q: %v", route, step, attr, selector, err)
		return
	}
	if got == nil || *got != want {
		var gotStr string
		if got != nil {
			gotStr = *got
		}
		t.Errorf("%s [%s]: %q[%s] = %q, want %q", route, step, selector, attr, gotStr, want)
	}
}

// expectNoConsoleErrors fails when the page logged a console error. External
// resources are not exempted here because a console error is the app's own
// JavaScript failing, not a network hiccup.
func expectNoConsoleErrors(t *testing.T, page *pageSession, route, step string) {
	t.Helper()
	if errs := page.ConsoleErrors(); len(errs) > 0 {
		var b strings.Builder
		for _, e := range errs {
			b.WriteString("\n  - ")
			b.WriteString(e.Text)
		}
		t.Errorf("%s [%s]: page logged %d console error(s):%s", route, step, len(errs), b.String())
	}
}

// expectNoBrokenSameOriginAssets fails when a subresource (image, stylesheet,
// script, font) returned ≥400 or failed to load. Two categories are exempt by
// design:
//
//   - external assets (the htmx CDN) — warnings only, per the plan default;
//   - the top-level document — an error route's 404 is the correct response and
//     is asserted by the render-contract tests, so it is not a "broken asset".
func expectNoBrokenSameOriginAssets(t *testing.T, page *pageSession, route, step string) {
	t.Helper()
	var broken []string
	for _, f := range page.FailedRequests() {
		if strings.Contains(f.Detail, "unpkg.com") {
			t.Logf("%s [%s]: external asset warning: %s", route, step, f.Detail)
			continue
		}
		if f.Type == "Document" {
			// The navigation's own status is the test's concern, not the asset
			// crawler's.
			continue
		}
		broken = append(broken, f.Detail)
	}
	if len(broken) > 0 {
		t.Errorf("%s [%s]: %d broken same-origin request(s):\n  - %s",
			route, step, len(broken), strings.Join(broken, "\n  - "))
	}
}
