package handlers

import (
	"net/http"
	"strings"
	"testing"
)

// crawlerEntryPoints are the pages whose same-origin links and assets are
// crawled. Together they reach every route and every asset reference in the
// app.
var crawlerEntryPoints = []string{
	"/",
	"/gallery",
	"/builders?zip=32801",
	"/builders?zip=10001",
	"/builders?zip=99999",
	"/builders/1",
	"/does-not-exist",
}

// ignoredHrefs are links the crawler does not follow. Each is a documented
// placeholder or non-navigable scheme, not a real route:
//   - "#"                     the "Estimate Costs" nav placeholder
//   - "javascript:..."        the profile "Back to Results" history link
//   - "mailto:" / "tel:"      contact links, not HTTP routes
func ignoredHrefs(href string) bool {
	return href == "#" ||
		strings.HasPrefix(href, "javascript:") ||
		strings.HasPrefix(href, "mailto:") ||
		strings.HasPrefix(href, "tel:")
}

// TestCrawl_InternalLinks extracts every same-origin <a href> reachable from
// the entry points and asserts no link 404s. This is the guard against broken
// navigation that only manifests when a user clicks.
func TestCrawl_InternalLinks(t *testing.T) {
	server := setupTestServer(t)

	seen := map[string]bool{}
	for _, entry := range crawlerEntryPoints {
		page := fetch(t, server, entry)
		if page.Status != 200 && page.Status != 404 {
			t.Fatalf("GET %s: unexpected status %d", entry, page.Status)
		}

		for _, a := range byTag(page.Doc, "a") {
			href := attr(a, "href")
			if href == "" || ignoredHrefs(href) {
				continue
			}
			// Same-origin only: skip absolute URLs to other hosts.
			if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
				continue
			}
			if !strings.HasPrefix(href, "/") {
				t.Errorf("GET %s: relative link %q is not same-origin-absolute", entry, href)
				continue
			}
			if seen[href] {
				continue
			}
			seen[href] = true

			resp, err := http.Get(server.URL + href)
			if err != nil {
				t.Errorf("GET %s (linked from %s): %v", href, entry, err)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusNotFound {
				t.Errorf("GET %s (linked from %s): link is dead (404)", href, entry)
			}
		}
	}

	if len(seen) == 0 {
		t.Fatal("no same-origin links discovered — crawler is not working")
	}
	t.Logf("crawled %d distinct links", len(seen))
}

// TestCrawl_StaticAssets extracts every /static/... reference (href and src)
// from every page and asserts it is served with 200. This is the guard against
// dead image paths, which are numerous and otherwise unverified.
func TestCrawl_StaticAssets(t *testing.T) {
	server := setupTestServer(t)

	seen := map[string]bool{}
	for _, entry := range crawlerEntryPoints {
		page := fetch(t, server, entry)

		for _, n := range elements(page.Doc) {
			var ref string
			switch n.Data {
			case "img":
				ref = attr(n, "src")
			case "link":
				ref = attr(n, "href")
			case "script":
				ref = attr(n, "src")
			default:
				continue
			}
			if !strings.HasPrefix(ref, "/static/") {
				continue
			}
			if seen[ref] {
				continue
			}
			seen[ref] = true

			resp, err := http.Get(server.URL + ref)
			if err != nil {
				t.Errorf("GET %s (referenced from %s): %v", ref, entry, err)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("GET %s (referenced from %s): status %d, want 200 (run `make css` if output.css is missing)", ref, entry, resp.StatusCode)
			}
		}
	}

	if len(seen) == 0 {
		t.Fatal("no static assets discovered — crawler is not working")
	}
	t.Logf("crawled %d distinct static assets", len(seen))
}

// TestCrawl_ImagesAreSameOrigin asserts every <img> points at a same-origin
// /static path. Cross-origin or relative image sources are a deployment bug.
func TestCrawl_ImagesAreSameOrigin(t *testing.T) {
	server := setupTestServer(t)

	var checked int
	for _, entry := range crawlerEntryPoints {
		page := fetch(t, server, entry)
		for _, img := range byTag(page.Doc, "img") {
			src := attr(img, "src")
			if src == "" {
				t.Errorf("GET %s: <img> with empty src", entry)
				continue
			}
			if !strings.HasPrefix(src, "/static/") {
				t.Errorf("GET %s: <img src=%q> is not a same-origin /static path", entry, src)
				continue
			}
			checked++
		}
	}
	t.Logf("checked %d <img> elements", checked)
}

// TestCrawl_RoutesRegistered is a lightweight tripwire: it asserts the exact
// route set the app exposes. Adding or removing a route without updating the
// crawler entry points should be a deliberate act.
func TestCrawl_RoutesRegistered(t *testing.T) {
	server := setupTestServer(t)

	cases := []struct {
		method, path string
		want         int
	}{
		{"GET", "/", 200},
		{"GET", "/gallery", 200},
		{"GET", "/builders", 200},
		{"GET", "/builders/1", 200},
		{"GET", "/api/builders?zip=32801", 200},
		{"GET", "/healthz", 200},
		{"GET", "/static/css/output.css", 200},
		{"GET", "/nonexistent", 404},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, server.URL+tc.path, nil)
			if err != nil {
				t.Fatalf("new request: %v", err)
			}
			resp, err := server.Client().Do(req)
			if err != nil {
				t.Fatalf("%s %s: %v", tc.method, tc.path, err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.want {
				t.Errorf("%s %s: status = %d, want %d", tc.method, tc.path, resp.StatusCode, tc.want)
			}
		})
	}
}
