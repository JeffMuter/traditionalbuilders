//go:build e2e

package e2e

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// axeSource is the pinned axe-core bundle. Vendoring it makes the a11y run
// reproducible and network-independent, and keeps the version explicit rather
// than whatever the CDN served that day. Bump it deliberately.
//
//go:embed vendor/axe.min.js
var axeSource string

// a11yRoutes are the pages scanned for serious/critical accessibility
// violations. A 404 is included because error pages are still user-facing.
var a11yRoutes = map[string]string{
	"/":                   "landing",
	"/gallery":            "gallery",
	"/builders?zip=32801": "builders",
	"/builders/1":         "profile",
	"/does-not-exist":     "error-404",
}

// axeViolation is the subset of axe's result we act on. Nodes carries the CSS
// selectors and HTML snippets of the offending elements so the failure message
// points at the actual markup.
type axeViolation struct {
	ID          string    `json:"id"`
	Impact      string    `json:"impact"`
	Description string    `json:"description"`
	Help        string    `json:"help"`
	Nodes       []axeNode `json:"nodes"`
}

type axeNode struct {
	Target []string `json:"target"`
	HTML   string   `json:"html"`
	// FailureSummary is axe's plain-language explanation for this node.
	FailureSummary string `json:"failureSummary"`
}

// TestAccessibility_NoSeriousOrCriticalViolations runs axe-core on every route
// and fails on serious/critical violations. Moderate/minor violations are
// logged as warnings so the suite is not blocked by low-signal noise.
func TestAccessibility_NoSeriousOrCriticalViolations(t *testing.T) {
	if axeSource == "" {
		t.Fatal("axe-core bundle is empty")
	}

	for route, name := range a11yRoutes {
		t.Run(name, func(t *testing.T) {
			page := newPage(t, "a11y-"+name)
			page.page.MustNavigate(routeURL(route)).MustWaitLoad()

			if err := page.page.AddScriptTag("", axeSource); err != nil {
				t.Fatalf("%s: inject axe-core: %v", route, err)
			}

			violations, err := runAxe(page)
			if err != nil {
				t.Fatalf("%s: run axe: %v", route, err)
			}

			var blocking []axeViolation
			for _, v := range violations {
				switch v.Impact {
				case "serious", "critical":
					blocking = append(blocking, v)
				default:
					t.Logf("%s: a11y warning [%s] %s", route, v.Impact, v.Help)
				}
			}
			if len(blocking) > 0 {
				var b strings.Builder
				for _, v := range blocking {
					fmt.Fprintf(&b, "\n  - [%s] %s: %s", v.Impact, v.ID, v.Help)
					for _, n := range v.Nodes {
						fmt.Fprintf(&b, "\n      target: %s", strings.Join(n.Target, " "))
						fmt.Fprintf(&b, "\n      html:   %s", n.HTML)
						if n.FailureSummary != "" {
							fmt.Fprintf(&b, "\n      why:    %s", strings.Join(strings.Fields(n.FailureSummary), " "))
						}
					}
				}
				t.Errorf("%s: %d serious/critical accessibility violation(s):%s",
					route, len(blocking), b.String())
			}
		})
	}
}

// runAxe evaluates axe.run() against the current document and returns the
// violations. The result is marshalled to JSON by the page to avoid duplicating
// axe's full type surface in Go.
func runAxe(page *pageSession) ([]axeViolation, error) {
	const js = `async () => {
		const results = await axe.run(document, {
			resultTypes: ['violations'],
		});
		return JSON.stringify(results.violations.map(v => ({
			id: v.id,
			impact: v.impact,
			description: v.description,
			help: v.help,
			nodes: v.nodes.map(n => ({
				target: n.target,
				html: n.html,
				failureSummary: n.failureSummary,
			})),
		})));
	}`

	raw, err := page.page.Eval(js)
	if err != nil {
		return nil, err
	}
	jsonStr := raw.Value.Str()
	if jsonStr == "" {
		return nil, fmt.Errorf("axe returned an empty result")
	}
	return decodeAxe(jsonStr)
}

// decodeAxe unmarshals the JSON violations emitted by the in-page axe run.
func decodeAxe(jsonStr string) ([]axeViolation, error) {
	var violations []axeViolation
	if err := json.Unmarshal([]byte(jsonStr), &violations); err != nil {
		return nil, fmt.Errorf("decode axe results: %w", err)
	}
	return violations, nil
}

// TestCapturePageScreenshots writes a screenshot of every key page on success.
// CI uploads this as a success artifact so a human can review the front-end
// from a PR without running anything.
func TestCapturePageScreenshots(t *testing.T) {
	dir := filepath.Join(repoRoot, "e2e", "artifacts", "screenshots")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create screenshots dir: %v", err)
	}

	for route, name := range hygieneRoutes {
		t.Run(name, func(t *testing.T) {
			page := newPage(t, "screenshot-"+name)
			page.page.MustNavigate(routeURL(route)).MustWaitLoad()
			// A short settle avoids capturing mid-transition paint.
			page.page.MustWaitStable()

			png, err := page.page.Screenshot(true, nil)
			if err != nil {
				t.Fatalf("%s: screenshot: %v", route, err)
			}
			out := filepath.Join(dir, name+".png")
			if err := os.WriteFile(out, png, 0o644); err != nil {
				t.Fatalf("%s: write screenshot: %v", route, err)
			}
			t.Logf("%s -> %s", route, out)
		})
	}
}
