//go:build e2e

package e2e

import (
	"testing"
)

// htmlHasDarkClass reports whether <html> currently carries the `dark` class.
func htmlHasDarkClass(t *testing.T, page *pageSession) bool {
	t.Helper()
	class, err := page.page.MustElement("html").Attribute("class")
	if err != nil || class == nil {
		return false
	}
	return containsToken(*class, "dark")
}

func containsToken(classList, token string) bool {
	for _, f := range splitFields(classList) {
		if f == token {
			return true
		}
	}
	return false
}

func splitFields(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

// storedTheme reads localStorage.theme, returning "" when unset.
func storedTheme(t *testing.T, page *pageSession) string {
	t.Helper()
	res, err := page.page.Eval(`() => localStorage.getItem('theme') ?? ""`)
	if err != nil {
		t.Fatalf("read localStorage.theme: %v", err)
	}
	return res.Value.Str()
}

// TestTheme_ToggleAndPersist asserts the theme toggle flips the <html> class,
// records the choice in localStorage, and survives a reload.
func TestTheme_ToggleAndPersist(t *testing.T) {
	page := newPage(t, "theme-toggle")
	page.page.MustNavigate(routeURL("/")).MustWaitLoad()

	// Start from a known light state.
	if _, err := page.page.Eval(`() => { localStorage.setItem('theme','light'); document.documentElement.classList.remove('dark') }`); err != nil {
		t.Fatalf("seed theme state: %v", err)
	}

	page.page.MustElement(`[data-testid="theme-toggle"]`).MustClick()

	if !htmlHasDarkClass(t, page) {
		t.Error("/: after toggle, <html> does not have the dark class")
	}
	if got := storedTheme(t, page); got != "dark" {
		t.Errorf("/: localStorage.theme = %q after toggle, want dark", got)
	}

	// Reload and confirm persistence (the inline head script re-applies it).
	page.page.MustReload().MustWaitLoad()
	if !htmlHasDarkClass(t, page) {
		t.Error("/: dark class did not persist across reload")
	}

	// Toggle back to light.
	page.page.MustElement(`[data-testid="theme-toggle"]`).MustClick()
	if htmlHasDarkClass(t, page) {
		t.Error("/: after second toggle, <html> still has the dark class")
	}
	if got := storedTheme(t, page); got != "light" {
		t.Errorf("/: localStorage.theme = %q after second toggle, want light", got)
	}
}
