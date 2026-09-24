package handlers

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// fullDocPages enumerates every route that must render the complete site
// chrome (doctype, head, nav header, footer, CSS link), with the status it is
// expected to return. It is the structural baseline every page shares.
var fullDocPages = map[string]struct {
	url    string
	status int
}{
	"landing":        {"/", 200},
	"gallery":        {"/gallery", 200},
	"builders":       {"/builders?zip=32801", 200},
	"builders-empty": {"/builders?zip=10001", 200},
	"builders-error": {"/builders?zip=99999", 200},
	"profile":        {"/builders/1", 200},
	"profile-404":    {"/builders/9999", 404},
	"error-404":      {"/does-not-exist", 404},
}

// TestRenderContract_SharedStructure asserts the elements every full page must
// contain. A regression here (missing CSS link, dropped footer, broken nav)
// breaks the whole site at once, so it is checked for every route.
func TestRenderContract_SharedStructure(t *testing.T) {
	server := setupTestServer(t)

	for name, tc := range fullDocPages {
		t.Run(name, func(t *testing.T) {
			url := tc.url
			page := fetch(t, server, url)
			expectStatus(t, page, tc.status)

			if ct := page.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
				t.Errorf("GET %s: Content-Type = %q, want text/html", url, ct)
			}
			if !strings.HasPrefix(strings.ToLower(page.Body), "<!doctype html>") {
				t.Errorf("GET %s: body does not start with doctype", url)
			}

			titles := byTag(page.Doc, "title")
			if len(titles) != 1 || textContent(titles[0]) == "" {
				t.Errorf("GET %s: want exactly one non-empty <title>, got %d", url, len(titles))
			}

			var cssFound bool
			for _, link := range byTag(page.Doc, "link") {
				if attr(link, "rel") == "stylesheet" && attr(link, "href") == "/static/css/output.css" {
					cssFound = true
				}
			}
			if !cssFound {
				t.Errorf("GET %s: missing <link rel=stylesheet href=/static/css/output.css>", url)
			}

			if len(byTag(page.Doc, "header")) == 0 {
				t.Errorf("GET %s: missing <header>", url)
			}
			if len(byTag(page.Doc, "footer")) == 0 {
				t.Errorf("GET %s: missing <footer>", url)
			}
			if byTestID(page.Doc, "theme-toggle") == nil {
				t.Errorf("GET %s: missing theme toggle", url)
			}
		})
	}
}

// TestRenderContract_ZipSearchForm pins the search form's contract. The input
// constraints matter: maxlength/pattern stop malformed zips client-side and
// required prevents an empty submit.
func TestRenderContract_ZipSearchForm(t *testing.T) {
	server := setupTestServer(t)
	page := fetch(t, server, "/")
	expectStatus(t, page, 200)

	form := byTestID(page.Doc, "zip-form")
	if form == nil {
		t.Fatal("landing: missing zip search form")
	}
	if got := attr(form, "action"); got != "/builders" {
		t.Errorf("form action = %q, want /builders", got)
	}
	if got := attr(form, "method"); !strings.EqualFold(got, "GET") {
		t.Errorf("form method = %q, want GET", got)
	}

	input := byTestID(page.Doc, "zip-search")
	if input == nil {
		t.Fatal("landing: missing zip input")
	}
	if got := attr(input, "name"); got != "zip" {
		t.Errorf("input name = %q, want zip", got)
	}
	if got := attr(input, "maxlength"); got != "5" {
		t.Errorf("input maxlength = %q, want 5", got)
	}
	if got := attr(input, "pattern"); got != `\d{5}` {
		t.Errorf("input pattern = %q, want %s", got, `\d{5}`)
	}
	if _, ok := hasAttr(input, "required"); !ok {
		t.Error("input is not required")
	}

	submit := byTestID(page.Doc, "zip-submit")
	if submit == nil {
		t.Fatal("landing: missing submit button")
	}
	if got := attr(submit, "type"); got != "submit" {
		t.Errorf("submit type = %q, want submit", got)
	}
}

func hasAttr(n *html.Node, name string) (string, bool) {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, name) {
			return a.Val, true
		}
	}
	return "", false
}

// TestRenderContract_BuildersStates asserts the three builders-page states are
// distinguishable by their testids: results, empty, error.
func TestRenderContract_BuildersStates(t *testing.T) {
	server := setupTestServer(t)

	t.Run("results", func(t *testing.T) {
		page := fetch(t, server, "/builders?zip=32801")
		expectStatus(t, page, 200)
		if byTestID(page.Doc, "results") == nil {
			t.Fatal("results state not rendered")
		}
		cards := byTestIDAll(page.Doc, "builder-card")
		if len(cards) == 0 {
			t.Fatal("results state has no builder cards")
		}
		for _, card := range cards {
			link := byTestID(card, "view-profile")
			if link == nil {
				t.Error("builder card missing view-profile link")
				continue
			}
			href := attr(link, "href")
			if !strings.HasPrefix(href, "/builders/") {
				t.Errorf("view-profile href = %q, want /builders/{id}", href)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		page := fetch(t, server, "/builders?zip=10001")
		expectStatus(t, page, 200)
		if byTestID(page.Doc, "empty-state") == nil {
			t.Error("empty state not rendered for valid zip with no builders")
		}
	})

	t.Run("unknown-zip-error", func(t *testing.T) {
		page := fetch(t, server, "/builders?zip=99999")
		expectStatus(t, page, 200)
		el := byTestID(page.Doc, "search-error")
		if el == nil {
			t.Fatal("error state not rendered for unknown zip")
		}
		if !strings.Contains(textContent(el), "recognize") {
			t.Errorf("error state copy = %q, want it to mention not recognizing the zip", textContent(el))
		}
	})

	t.Run("malformed-zip-error", func(t *testing.T) {
		page := fetch(t, server, "/builders?zip=abc")
		expectStatus(t, page, 200)
		el := byTestID(page.Doc, "search-error")
		if el == nil {
			t.Fatal("error state not rendered for malformed zip")
		}
		if !strings.Contains(textContent(el), "valid 5-digit zip code") {
			t.Errorf("error state copy = %q, want validation message", textContent(el))
		}
	})
}

// TestRenderContract_Profile asserts a known builder renders their name and the
// projects section, and unknown/malformed ids render a styled 404.
func TestRenderContract_Profile(t *testing.T) {
	server := setupTestServer(t)

	t.Run("known", func(t *testing.T) {
		page := fetch(t, server, "/builders/1")
		expectStatus(t, page, 200)
		if byTestID(page.Doc, "profile") == nil {
			t.Fatal("profile container not rendered")
		}
		if byTestID(page.Doc, "projects") == nil {
			t.Error("projects section not rendered")
		}
		if !strings.Contains(textContent(page.Doc), "Near Orlando (closest)") {
			t.Error("profile does not render the builder name")
		}
	})

	for _, tc := range []struct{ name, url string }{
		{"non-numeric id", "/builders/abc"},
		{"unknown id", "/builders/9999"},
		{"zero id", "/builders/0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			page := fetch(t, server, tc.url)
			expectStatus(t, page, 404)
			el := byTestID(page.Doc, "profile-not-found")
			if el == nil {
				t.Fatal("ProfileNotFound body not rendered")
			}
			if !strings.Contains(textContent(el), "couldn't find that builder") {
				t.Errorf("not-found copy = %q", textContent(el))
			}
		})
	}
}

// TestRenderContract_ErrorPage asserts unmatched routes get the styled 404 and
// /healthz reports readiness.
func TestRenderContract_ErrorPage(t *testing.T) {
	server := setupTestServer(t)

	page := fetch(t, server, "/does-not-exist")
	expectStatus(t, page, 404)
	if byTestID(page.Doc, "error-page") == nil {
		t.Fatal("styled ErrorPage body not rendered")
	}
	if !strings.Contains(textContent(page.Doc), "Page not found") {
		t.Errorf("error page copy = %q", textContent(page.Doc))
	}

	resp, err := server.Client().Get(server.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("/healthz status = %d, want 200", resp.StatusCode)
	}
}

// TestRenderContract_NavPlaceholder documents the known dead nav link. The
// "Estimate Costs" link points at "#" today; the link crawler allowlists it.
// When the estimate-costs feature ships, this test should be updated to assert
// the real route so the placeholder cannot linger unnoticed.
func TestRenderContract_NavPlaceholder(t *testing.T) {
	server := setupTestServer(t)
	page := fetch(t, server, "/")
	expectStatus(t, page, 200)

	var found bool
	for _, a := range byTag(page.Doc, "a") {
		if attr(a, "href") == "#" && strings.Contains(textContent(a), "Estimate Costs") {
			found = true
		}
	}
	if !found {
		t.Error("expected the known 'Estimate Costs' href=\"#\" placeholder; if it moved to a real route, update this test and the crawler allowlist")
	}
}
